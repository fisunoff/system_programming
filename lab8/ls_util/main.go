package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unsafe"
)

// Опции командной строки
var (
	showAttributes bool // -L
	recursive      bool // -R
)

func init() {
	flag.BoolVar(&showAttributes, "L", false, "Показывать атрибуты файлов")
	flag.BoolVar(&recursive, "R", false, "Рекурсивный обход подкаталогов")
}

func main() {
	flag.Parse()

	patterns := flag.Args()
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}

	for _, pattern := range patterns {
		err := processPattern(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при обработке %s: %v\n", pattern, err)
		}
	}
}

func processPattern(pattern string) error {
	dir, mask := filepath.Split(pattern)
	if dir == "" {
		dir = "."
	}

	regexPattern := convertMaskToRegex(mask)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return fmt.Errorf("неверный шаблон %s: %v", mask, err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	if recursive {
		return filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if re.MatchString(info.Name()) {
				printFileInfo(path, info)
			}
			return nil
		})
	}

	return listDirectory(absDir, re)
}

func convertMaskToRegex(mask string) string {
	mask = strings.ReplaceAll(mask, "\\", "\\\\")
	regex := regexp.QuoteMeta(mask)
	regex = strings.ReplaceAll(regex, "\\*", ".*")
	regex = strings.ReplaceAll(regex, "\\?", ".")
	return "^" + regex + "$"
}

func printFileInfo(path string, info os.FileInfo) {
	name := info.Name()

	// Время изменения
	timeStr := info.ModTime().Format("2006-01-02 15:04:05")

	// РАЗМЕР ФАЙЛА в человеко-читаемом формате
	size := info.Size()
	sizeStr := formatSize(size)

	// Младшая часть размера (последние 6 цифр) - как требуется в задании
	lowPartStr := fmt.Sprintf("%06d", size%1000000)

	// Определяем тип
	fileType := "-"
	if info.IsDir() {
		fileType = "d"
	}

	// Выводим информацию с размером
	fmt.Printf("%s %s %s %s %s",
		fileType,
		timeStr,
		sizeStr,    // человеко-читаемый размер
		lowPartStr, // младшая часть размера
		name)

	if info.IsDir() {
		fmt.Print("\\")
	}
	fmt.Println()

	if showAttributes {
		printAttributes(path, info)
	}
}

// formatSize форматирует размер в человеко-читаемый вид
func formatSize(size int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%6.2fTB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%6.2fGB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%6.2fMB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%6.2fKB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%7dB", size)
	}
}

func printAttributes(path string, info os.FileInfo) {
	// Пытаемся получить атрибуты через Windows API
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return
	}

	var attrs uint32
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getFileAttributes := kernel32.NewProc("GetFileAttributesW")

	ret, _, _ := getFileAttributes.Call(uintptr(unsafe.Pointer(pathPtr)))
	if ret != 0xFFFFFFFF {
		attrs = uint32(ret)
	}

	attrStr := formatWindowsAttributes(attrs)
	fmt.Printf("  [%s] Размер: %d байт\n", attrStr, info.Size())
}

func formatWindowsAttributes(attr uint32) string {
	var attrs []string

	if attr&0x00000001 != 0 {
		attrs = append(attrs, "R")
	} // READONLY
	if attr&0x00000002 != 0 {
		attrs = append(attrs, "H")
	} // HIDDEN
	if attr&0x00000004 != 0 {
		attrs = append(attrs, "S")
	} // SYSTEM
	if attr&0x00000010 != 0 {
		attrs = append(attrs, "D")
	} // DIRECTORY
	if attr&0x00000020 != 0 {
		attrs = append(attrs, "A")
	} // ARCHIVE

	if len(attrs) == 0 {
		return "---"
	}
	return strings.Join(attrs, "")
}

func listDirectory(path string, re *regexp.Regexp) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()

	entries, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	// Выводим информацию о каталоге
	fmt.Printf("\nСодержимое каталога %s:\n", path)
	fmt.Println(strings.Repeat("-", 60))

	for _, entry := range entries {
		if re.MatchString(entry.Name()) {
			fullPath := filepath.Join(path, entry.Name())
			printFileInfo(fullPath, entry)
		}
	}

	return nil
}
