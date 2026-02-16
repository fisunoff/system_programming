package main

import (
	"fmt"
	"strings"
	"syscall"
	"time"
)

// Глобальный флаг выхода
var exitFlag bool = false

// Константы для обработчиков событий консоли Windows
const (
	CTRL_C_EVENT        = 0 // Ctrl+C
	CTRL_BREAK_EVENT    = 1 // Ctrl+Break
	CTRL_CLOSE_EVENT    = 2 // Закрытие окна консоли
	CTRL_LOGOFF_EVENT   = 5 // Выход пользователя из системы
	CTRL_SHUTDOWN_EVENT = 6 // Завершение работы системы
)

// HandlerFunc тип функции обработчика
type HandlerFunc func(event uint32) bool

var handlerFunc HandlerFunc

// Структура для хранения информации о событии
type EventInfo struct {
	Name        string
	Description string
	IsSystem    bool
}

// Маппинг событий в понятные описания
var eventNames = map[uint32]EventInfo{
	CTRL_C_EVENT: {
		Name:        "CTRL_C_EVENT",
		Description: "Нажатие Ctrl+C",
		IsSystem:    false,
	},
	CTRL_BREAK_EVENT: {
		Name:        "CTRL_BREAK_EVENT",
		Description: "Нажатие Ctrl+Break",
		IsSystem:    false,
	},
	CTRL_CLOSE_EVENT: {
		Name:        "CTRL_CLOSE_EVENT",
		Description: "Закрытие окна консоли",
		IsSystem:    false,
	},
	CTRL_LOGOFF_EVENT: {
		Name:        "CTRL_LOGOFF_EVENT",
		Description: "Выход пользователя из системы",
		IsSystem:    true,
	},
	CTRL_SHUTDOWN_EVENT: {
		Name:        "CTRL_SHUTDOWN_EVENT",
		Description: "Завершение работы системы",
		IsSystem:    true,
	},
}

// Beep вызывает стандартный звуковой сигнал через Windows API
func beep() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	beepProc := kernel32.NewProc("Beep")
	beepProc.Call(440, 200) // 440 Гц, 200 мс
}

// systemBeep специальный звук для системных событий
func systemBeep() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	beepProc := kernel32.NewProc("Beep")
	beepProc.Call(880, 100) // Высокий тон
	beepProc.Call(440, 100) // Низкий тон
	beepProc.Call(880, 100)
}

//export consoleHandler
func consoleHandler(event uint32) int {
	// Вызываем нашу функцию-обработчик
	if handlerFunc != nil {
		if handlerFunc(event) {
			return 1 // TRUE - событие обработано
		}
	}
	return 0 // FALSE - передать событие следующему обработчику
}

func main() {
	fmt.Println("=== Программа обработки сигналов консоли Windows ===")
	fmt.Println("Версия 2.0 - с поддержкой системных событий")
	fmt.Println()
	fmt.Println("Программа Ctrlc запущена. Нажмите Ctrl+C для завершения...")
	fmt.Println("Также обрабатываются:")
	fmt.Println("  - Ctrl+Break")
	fmt.Println("  - Закрытие окна консоли")
	fmt.Println("  - Выход пользователя из системы")
	fmt.Println("  - Завершение работы системы")
	fmt.Println()

	// Устанавливаем обработчик событий консоли
	err := setConsoleCtrlHandler(consoleHandler, true)
	if err != nil {
		fmt.Printf("Ошибка установки обработчика: %v\n", err)
		return
	}

	// Устанавливаем нашу функцию-обработчик
	handlerFunc = handleConsoleEvent

	fmt.Println("Обработчик событий консоли установлен.")
	fmt.Println()

	// Бесконечный цикл с проверкой флага выхода
	counter := 0
	for !exitFlag {
		counter++

		// Вызываем Beep каждые 5 секунд
		currentTime := time.Now().Format("15:04:05")
		fmt.Printf("[%s] Работаем... Beep! (цикл #%d)\n", currentTime, counter)
		beep()

		// Ждем 5 секунд с проверкой флага
		for i := 0; i < 5; i++ {
			if exitFlag {
				break
			}
			time.Sleep(1 * time.Second)

			if !exitFlag && i < 4 {
				fmt.Printf("[%s] До следующего сигнала: %d сек\n",
					time.Now().Format("15:04:05"), 4-i)
			}
		}
	}

	fmt.Println("\nПрограмма завершена по флагу выхода.")
}

// handleConsoleEvent обрабатывает события консоли
func handleConsoleEvent(event uint32) bool {
	// Получаем информацию о событии
	eventInfo, exists := eventNames[event]
	if !exists {
		eventInfo = EventInfo{
			Name:        fmt.Sprintf("UNKNOWN_EVENT_%d", event),
			Description: fmt.Sprintf("Неизвестное событие (%d)", event),
			IsSystem:    false,
		}
	}

	// Форматируем сообщение
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("[Обработчик] ПОЛУЧЕНО СОБЫТИЕ: %s\n", eventInfo.Name)
	fmt.Printf("[Обработчик] Описание: %s\n", eventInfo.Description)
	if eventInfo.IsSystem {
		fmt.Printf("[Обработчик] ВНИМАНИЕ: Это системное событие!\n")
	}
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	// Специальная обработка для системных событий
	if eventInfo.IsSystem {
		fmt.Println("[Обработчик] Системное событие - требуется особая обработка")
		systemBeep()
	}

	// Устанавливаем флаг выхода
	exitFlag = true
	fmt.Printf("[Обработчик] Флаг выхода установлен в %v\n", exitFlag)

	// Этап 1: Ожидание 4 секунды
	fmt.Println("[Обработчик] Фаза 1: Ожидание 4 секунды...")
	for i := 0; i < 4; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("[Обработчик] Прошла %d секунда (фаза 1)\n", i+1)
	}

	// Этап 2: Ожидание еще 6 секунд
	fmt.Println("[Обработчик] Фаза 2: Ожидание еще 6 секунд...")
	for i := 0; i < 6; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("[Обработчик] Прошло %d секунд (фаза 2)\n", i+5)
	}

	// Финальное сообщение
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	fmt.Println("[Обработчик] Завершение работы обработчика")
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	// Для системных событий делаем дополнительную паузу
	if eventInfo.IsSystem {
		fmt.Println("[Обработчик] Дополнительная пауза для системных событий...")
		time.Sleep(2 * time.Second)
	}

	return true // Возвращаем TRUE - событие обработано
}

// setConsoleCtrlHandler устанавливает обработчик событий консоли через Windows API
func setConsoleCtrlHandler(handler interface{}, add bool) error {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleCtrlHandler := kernel32.NewProc("SetConsoleCtrlHandler")

	// Получаем указатель на функцию
	handlerPtr := syscall.NewCallback(handler)

	ret, _, err := setConsoleCtrlHandler.Call(
		handlerPtr,
		boolToUintptr(add),
	)

	if ret == 0 {
		return err
	}
	return nil
}

// boolToUintptr преобразует bool в uintptr для Windows API
func boolToUintptr(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
