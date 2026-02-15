package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	STD_INPUT_HANDLE  = 0xFFFFFFF6 // -10
	STD_OUTPUT_HANDLE = 0xFFFFFFF5 // -11
	ENABLE_ECHO_INPUT = 0x0004
)

// kernel32 DLL для функций работы с режимом консоли (в syscall их может не быть напрямую)
var (
	modkernel32        = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = modkernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = modkernel32.NewProc("SetConsoleMode")
)

// PrintMsg Выводит одну строку в дескриптор.
func PrintMsg(hOut syscall.Handle, msg string) bool {
	b := []byte(msg)
	var written uint32
	err := syscall.WriteFile(hOut, b, &written, nil)
	return err == nil
}

// PrintStrings Выводит список строк переменной длины.
func PrintStrings(hOut syscall.Handle, msg ...string) bool {
	for _, m := range msg {
		if !PrintMsg(hOut, m) {
			return false
		}
	}
	return true
}

// ConsolePrompt Выводит приглашение и читает ответ.
func ConsolePrompt(pPromptMsg string, pResponse []byte, echo bool) (int, bool) {
	hOut, _ := syscall.GetStdHandle(STD_OUTPUT_HANDLE)
	hIn, _ := syscall.GetStdHandle(STD_INPUT_HANDLE)

	if !PrintMsg(hOut, pPromptMsg) {
		return 0, false
	}

	// настройка режима эха
	var originalMode uint32
	r1, _, _ := procGetConsoleMode.Call(uintptr(hIn), uintptr(unsafe.Pointer(&originalMode)))
	if r1 == 0 {
		return 0, false // ошибка получения режима
	}

	currentMode := originalMode
	if !echo {
		// Отключаем ECHO_INPUT
		currentMode &^= ENABLE_ECHO_INPUT
		procSetConsoleMode.Call(uintptr(hIn), uintptr(currentMode))
	}

	// Читаем ввод (ReadFile)
	var read uint32
	err := syscall.ReadFile(hIn, pResponse, &read, nil)

	// Восстанавливаем режим сразу после чтения
	if !echo {
		procSetConsoleMode.Call(uintptr(hIn), uintptr(originalMode))
	}

	if err != nil {
		return 0, false
	}

	// Убираем символы перевода строки (\r\n) из конца, если они есть
	count := int(read)
	if count > 0 && pResponse[count-1] == '\n' {
		count--
	}
	if count > 0 && pResponse[count-1] == '\r' {
		count--
	}

	return count, true
}

func main() {
	// Получаем стандартный дескриптор вывода для тестов
	hOut, err := syscall.GetStdHandle(STD_OUTPUT_HANDLE)
	if err != nil {
		fmt.Println("Error getting StdOut handle:", err)
		return
	}

	// ТЕСТ 1: PrintMsg
	PrintMsg(hOut, "--- Test 1: PrintMsg ---\r\n")
	PrintMsg(hOut, "Hello from single message!\r\n\r\n")

	// ТЕСТ 2: PrintStrings (аналог va_list)
	PrintMsg(hOut, "--- Test 2: PrintStrings ---\r\n")
	PrintStrings(hOut,
		"Line 1: This is ",
		"constructed from ",
		"multiple arguments.\r\n",
		"Line 2: Variadic functions in Go are cool.\r\n\r\n",
	)

	// ТЕСТ 3: ConsolePrompt с Эхо (обычный ввод)
	PrintMsg(hOut, "--- Test 3: ConsolePrompt (Echo ON) ---\r\n")
	buffer := make([]byte, 256) // MaxTchar = 256

	lenRead, ok := ConsolePrompt("Enter your name: ", buffer, true)
	if ok {
		name := string(buffer[:lenRead])
		PrintStrings(hOut, "Hello, ", name, "!\r\n\r\n")
	} else {
		PrintMsg(hOut, "Error reading name.\r\n")
	}

	// ТЕСТ 4: ConsolePrompt БЕЗ Эхо (пароль)
	PrintMsg(hOut, "--- Test 4: ConsolePrompt (Echo OFF / Password) ---\r\n")

	// Ввод пароля
	lenPass1, _ := ConsolePrompt("Create password: ", buffer, false)
	pass1 := string(buffer[:lenPass1])

	// Подтверждение пароля
	buffer2 := make([]byte, 256)
	lenPass2, _ := ConsolePrompt("Confirm password: ", buffer2, false)
	pass2 := string(buffer2[:lenPass2])

	// Проверка
	if pass1 == pass2 {
		PrintStrings(hOut, "Success: Passwords match!\r\n")
	} else {
		PrintStrings(hOut, "Error: Passwords do not match.\r\n")
		PrintStrings(hOut, "Got: '", pass1, "' and '", pass2, "'\r\n")
	}
}
