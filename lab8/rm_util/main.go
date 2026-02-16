package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Опции командной строки
var (
	recursive   bool // -r для рекурсивного удаления
	force       bool // -f для игнорирования несуществующих файлов
	verbose     bool // -v для подробного вывода
	interactive bool // -i для запроса подтверждения
)

func init() {
	flag.BoolVar(&recursive, "r", false, "Рекурсивное удаление каталогов")
	flag.BoolVar(&force, "f", false, "Игнорировать несуществующие файлы, не запрашивать подтверждение")
	flag.BoolVar(&verbose, "v", false, "Подробный вывод")
	flag.BoolVar(&interactive, "i", false, "Запрашивать подтверждение перед удалением")
}

func main() {
	flag.Parse()

	// Получаем имена файлов для удаления
	files := flag.Args()
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "rm: пропущен операнд\n")
		fmt.Fprintf(os.Stderr, "Использование: rm [-rRfvi] файл...\n")
		os.Exit(1)
	}

	exitCode := 0

	// Обрабатываем каждый файл
	for _, file := range files {
		err := removeItem(file)
		if err != nil {
			// Если включен force, игнорируем ошибки о несуществующих файлах
			if force && os.IsNotExist(err) {
				if verbose {
					fmt.Printf("rm: '%s' не существует (игнорируется -f)\n", file)
				}
				continue
			}
			fmt.Fprintf(os.Stderr, "rm: %v\n", err)
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

// removeItem пытается удалить элемент, сначала проверяя его существование
func removeItem(path string) error {
	// Сначала проверяем существование файла с помощью os.Stat
	// но перехватываем ошибку "не существует" для обработки с флагом -f
	info, err := os.Lstat(path)
	if err != nil {
		// Если файл не существует и включен force - это не ошибка
		if os.IsNotExist(err) {
			return err // Вернем ошибку, но main обработает её с учётом force
		}
		// Другие ошибки (нет прав доступа и т.д.) всегда показываем
		return fmt.Errorf("не удалось получить информацию о '%s': %v", path, err)
	}

	// Проверяем, является ли элемент символической ссылкой
	isSymlink := info.Mode()&os.ModeSymlink != 0

	// Если это каталог и не символическая ссылка
	if info.IsDir() && !isSymlink {
		return removeDirectory(path)
	}

	// Это файл или символическая ссылка
	return removeFile(path)
}

// removeDirectory обрабатывает удаление каталога
func removeDirectory(path string) error {
	// Если не рекурсивный режим, отказываемся удалять каталог
	if !recursive {
		return fmt.Errorf("не удалось удалить '%s': это каталог (используйте -r для рекурсивного удаления)", path)
	}

	// Запрашиваем подтверждение для рекурсивного удаления
	if interactive && !force {
		if !confirm(fmt.Sprintf("rm: удалить каталог '%s'? ", path)) {
			if verbose {
				fmt.Printf("rm: пропуск '%s'\n", path)
			}
			return nil
		}
	}

	// Читаем содержимое каталога
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("не удалось прочитать каталог '%s': %v", path, err)
	}

	// Рекурсивно удаляем содержимое
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		err := removeItem(entryPath)
		if err != nil {
			// Если файл не существует и включен force, пропускаем
			if force && os.IsNotExist(err) {
				continue
			}
			return err
		}
	}

	// Удаляем сам каталог
	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("не удалось удалить каталог '%s': %v", path, err)
	}

	if verbose {
		fmt.Printf("rm: удален каталог '%s'\n", path)
	}

	return nil
}

// removeFile обрабатывает удаление обычного файла или символической ссылки
func removeFile(path string) error {
	// Запрашиваем подтверждение если нужно
	if interactive && !force {
		if !confirm(fmt.Sprintf("rm: удалить '%s'? ", path)) {
			if verbose {
				fmt.Printf("rm: пропуск '%s'\n", path)
			}
			return nil
		}
	}

	// Удаляем файл
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("не удалось удалить '%s': %v", path, err)
	}

	if verbose {
		fmt.Printf("rm: удален '%s'\n", path)
	}

	return nil
}

// confirm запрашивает подтверждение у пользователя
func confirm(prompt string) bool {
	fmt.Print(prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(response)
	return response == "y" || response == "yes" || response == "д" || response == "да"
}

// removeWithPattern удаляет файлы по шаблону (дополнительная функция)
func removeWithPattern(pattern string) error {
	// Разделяем путь и маску
	dir, mask := filepath.Split(pattern)
	if dir == "" {
		dir = "."
	}

	// Проверяем существование каталога
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) && force {
			return nil // Если каталог не существует и включен force, игнорируем
		}
		return err
	}

	// Читаем содержимое каталога
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Проходим по всем записям
	for _, entry := range entries {
		// Проверяем соответствует ли имя маске
		matched, err := filepath.Match(mask, entry.Name())
		if err != nil {
			continue
		}
		if matched {
			fullPath := filepath.Join(dir, entry.Name())
			err := removeItem(fullPath)
			if err != nil {
				// Если файл не существует и включен force, пропускаем
				if force && os.IsNotExist(err) {
					continue
				}
				fmt.Fprintf(os.Stderr, "rm: %v\n", err)
			}
		}
	}

	return nil
}
