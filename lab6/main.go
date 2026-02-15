package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Для хранения имен, переданных через аргументы
var (
	Name1 string
	Name2 string
)

// printHelp выводит справку
func printHelp() {
	fmt.Println("=== TestFileDir Help ===")
	fmt.Println("Program demonstrating file and directory operations.")
	fmt.Println("Usage: TestFileDir [Name1] [Name2]")
	fmt.Println("       TestFileDir /?")
	fmt.Println("Author: Фисунов Антон")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println(" - Create file/directory")
	fmt.Println(" - Delete file/directory")
	fmt.Println(" - Copy file")
	fmt.Println(" - Rename/Move file/directory")
}

// askConfirmation запрашивает подтверждение (Y/N)
func askConfirmation(action string) bool {
	fmt.Printf("Confirm action '%s'? (Y/N): ", action)
	reader := bufio.NewReader(os.Stdin)
	char, _, _ := reader.ReadRune()
	if char != '\n' {
		reader.ReadString('\n')
	}
	return char == 'y' || char == 'Y'
}

// checkExists проверяет существование файла/папки
func checkExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// copyFile копирует содержимое файла src в dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// actionCreateFile реализует Create File (Name1)
func actionCreateFile() error {
	if checkExists(Name1) {
		return fmt.Errorf("file '%s' already exists", Name1)
	}
	file, err := os.Create(Name1)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Printf("File '%s' created successfully.\n", Name1)
	return nil
}

// actionCreateDir реализует Create Directory (Name1)
func actionCreateDir() error {
	if checkExists(Name1) {
		return fmt.Errorf("directory '%s' already exists", Name1)
	}
	// 0755 - стандартные права доступа (rwxr-xr-x)
	err := os.Mkdir(Name1, 0755)
	if err != nil {
		return err
	}
	fmt.Printf("Directory '%s' created successfully.\n", Name1)
	return nil
}

// actionDelete Delete (Name1)
func actionDelete() error {
	if !checkExists(Name1) {
		return fmt.Errorf("'%s' does not exist", Name1)
	}
	if !askConfirmation("Delete " + Name1) {
		return fmt.Errorf("operation cancelled by user")
	}

	// os.RemoveAll удаляет и файл, и непустую директорию
	err := os.RemoveAll(Name1)
	if err != nil {
		return err
	}
	fmt.Printf("'%s' deleted successfully.\n", Name1)
	return nil
}

// actionCopy Copy File (Name1 -> Name2)
func actionCopy() error {
	if !checkExists(Name1) {
		return fmt.Errorf("source '%s' does not exist", Name1)
	}
	if checkExists(Name2) {
		if !askConfirmation(fmt.Sprintf("Overwrite '%s'", Name2)) {
			return fmt.Errorf("operation cancelled")
		}
	}

	// Проверяем, является ли Name1 директорией (копирование папок сложнее,
	// реализуем только для файлов, как в задании "копировать файлы")
	info, err := os.Stat(Name1)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("source '%s' is a directory (copying dirs not implemented)", Name1)
	}

	err = copyFile(Name1, Name2)
	if err != nil {
		return err
	}
	fmt.Printf("Copied '%s' to '%s'.\n", Name1, Name2)
	return nil
}

// actionRename Rename / Move (Name1 -> Name2)
func actionRename() error {
	if !checkExists(Name1) {
		return fmt.Errorf("source '%s' does not exist", Name1)
	}
	// Если целевое имя существует - спрашиваем
	if checkExists(Name2) {
		return fmt.Errorf("destination '%s' already exists", Name2)
	}

	if !askConfirmation(fmt.Sprintf("Rename '%s' to '%s'", Name1, Name2)) {
		return fmt.Errorf("operation cancelled")
	}

	err := os.Rename(Name1, Name2)
	if err != nil {
		return err
	}
	fmt.Printf("Renamed '%s' to '%s'.\n", Name1, Name2)

	// Обновляем имена в памяти, чтобы меню показывало актуальные данные
	// (Опционально, но логично)
	Name1 = Name2
	Name2 = ""

	return nil
}

func main() {
	args := os.Args[1:]

	if len(args) > 0 && (args[0] == "/?" || args[0] == "-h" || args[0] == "--help") {
		printHelp()
		return
	}

	if len(args) < 2 {
		fmt.Println("Error: Please provide two names (files or directories).")
		fmt.Println("Usage: TestFileDir Name1 Name2")
		fmt.Println("Try 'TestFileDir /?' for help.")
		return
	}

	Name1 = args[0]
	Name2 = args[1]

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n========================================")
		fmt.Printf(" Current Names: [1]: %s   [2]: %s\n", Name1, Name2)
		fmt.Println("========================================")
		fmt.Println("1. Create File (Name1)")
		fmt.Println("2. Create Directory (Name1)")
		fmt.Println("3. Delete (Name1)")
		fmt.Println("4. Copy File (Name1 -> Name2)")
		fmt.Println("5. Rename/Move (Name1 -> Name2)")
		fmt.Println("0. Exit")
		fmt.Println("========================================")
		fmt.Print("Select option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var err error

		switch input {
		case "1":
			err = actionCreateFile()
		case "2":
			err = actionCreateDir()
		case "3":
			err = actionDelete()
		case "4":
			err = actionCopy()
		case "5":
			err = actionRename()
		case "0":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
			continue
		}

		if err != nil {
			fmt.Printf("\n[ERROR]: %v\n", err)
		} else {
			fmt.Println("\n[SUCCESS]")
		}
	}
}
