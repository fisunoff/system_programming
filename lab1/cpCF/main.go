package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: cpCF source_file destination_file")
		return
	}

	src := os.Args[1]
	dst := os.Args[2]

	fmt.Printf("Copying %s to %s via Win32 CopyFile...\n", src, dst)

	// Загружаем kernel32.dll
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procCopyFile := kernel32.NewProc("CopyFileW")

	srcPtr, _ := syscall.UTF16PtrFromString(src)
	dstPtr, _ := syscall.UTF16PtrFromString(dst)

	// Параметр bFailIfExists: 0 (FALSE) - перезаписать, если существует
	failIfExists := uintptr(0)

	// Вызываем функцию.
	// CopyFileW(LPCWSTR lpExistingFileName, LPCWSTR lpNewFileName, BOOL bFailIfExists)
	ret, _, err := procCopyFile.Call(
		uintptr(unsafe.Pointer(srcPtr)),
		uintptr(unsafe.Pointer(dstPtr)),
		failIfExists,
	)

	// CopyFile возвращает не 0 при успехе
	if ret == 0 {
		fmt.Printf("Error copying file: %v\n", err)
	} else {
		fmt.Println("Success.")
	}
}
