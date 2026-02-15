package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: cpW source_file destination_file")
		return
	}

	srcPath, _ := syscall.UTF16PtrFromString(os.Args[1])
	destPath, _ := syscall.UTF16PtrFromString(os.Args[2])

	fmt.Printf("Copying %s to %s via Win32 CreateFile...\n", os.Args[1], os.Args[2])

	hSrc, err := syscall.CreateFile(
		srcPath,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		fmt.Printf("Error opening source: %v\n", err)
		return
	}
	defer syscall.CloseHandle(hSrc)

	hDst, err := syscall.CreateFile(
		destPath,
		syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.CREATE_ALWAYS,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)

	if err != nil {
		fmt.Printf("Error creating dest: %v\n", err)
		return
	}
	defer syscall.CloseHandle(hDst)

	var done uint32
	var written uint32
	buf := make([]byte, 4096)

	for {
		// ReadFile
		err = syscall.ReadFile(hSrc, buf, &done, nil)
		if err != nil && err != syscall.ERROR_HANDLE_EOF {
			fmt.Printf("Error reading: %v\n", err)
			return
		}
		if done == 0 {
			break // EOF
		}

		// WriteFile
		err = syscall.WriteFile(hDst, buf[:done], &written, nil)
		if err != nil {
			fmt.Printf("Error writing: %v\n", err)
			return
		}
	}

	fmt.Println("Success.")
}
