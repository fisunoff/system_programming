package main

import (
	"fmt"
	"io"
	"os"
)

func CatFile(src io.Reader, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	return err
}

func main() {
	args := os.Args[1:]
	suppressErrors := false
	var files []string

	for _, arg := range args {
		if arg == "-s" {
			suppressErrors = true
			continue
		}
		// Все, что не флаг - считаем файлами
		files = append(files, arg)
	}

	// Если файлов нет — читаем Stdin
	if len(files) == 0 {
		if err := CatFile(os.Stdin, os.Stdout); err != nil {
			// Ошибки вывода в Stderr
			if !suppressErrors {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		}
		return
	}

	// Читаем файлы по очереди
	for _, fname := range files {
		// Поддержка чтения из stdin через -
		if fname == "-" {
			err := CatFile(os.Stdin, os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			continue
		}

		f, err := os.Open(fname)
		if err != nil {
			if !suppressErrors {
				fmt.Fprintf(os.Stderr, "cat: cannot open %s: %v\n", fname, err)
			}
			continue
		}

		if err := CatFile(f, os.Stdout); err != nil {
			if !suppressErrors {
				fmt.Fprintf(os.Stderr, "cat: error reading %s: %v\n", fname, err)
			}
		}
		f.Close()
	}
}
