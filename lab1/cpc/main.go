package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <errno.h>

// Обертка для копирования, так как макросы и указатели FILE* удобнее обрабатывать в блоке C
int copy_file(char* srcPath, char* dstPath) {
    FILE *src = fopen(srcPath, "rb");
    if (src == NULL) return -1;

    FILE *dst = fopen(dstPath, "wb");
    if (dst == NULL) {
        fclose(src);
        return -2;
    }

    char buffer[4096];
    size_t bytesRead;

    while ((bytesRead = fread(buffer, 1, sizeof(buffer), src)) > 0) {
        fwrite(buffer, 1, bytesRead, dst);
    }

    fclose(src);
    fclose(dst);
    return 0;
}
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: cpC source_file destination_file")
		return
	}

	srcFile := C.CString(os.Args[1])
	dstFile := C.CString(os.Args[2])

	// Освобождаем память C-строк после завершения работы
	defer C.free(unsafe.Pointer(srcFile))
	defer C.free(unsafe.Pointer(dstFile))

	fmt.Printf("Copying %s to %s via C stdio...\n", os.Args[1], os.Args[2])

	res := C.copy_file(srcFile, dstFile)
	if res == 0 {
		fmt.Println("Success.")
	} else {
		fmt.Printf("Error occurred. Code: %d\n", res)
	}
}
