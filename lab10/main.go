// regEdit.go
package main

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func main() {
	fmt.Println("=== Интерактивный редактор реестра ===")
	fmt.Println()

	for {
		fmt.Print("Введите путь к разделу реестра (или 'exit' для выхода): ")
		var regPath string
		fmt.Scanln(&regPath)

		if strings.ToLower(regPath) == "exit" {
			break
		}

		fmt.Print("Введите имя параметра: ")
		var paramName string
		fmt.Scanln(&paramName)

		key, err := openRegistryKeyForEdit(regPath)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}

		currentValue, valueType, err := readRegistryValue(key, paramName)
		key.Close()

		if err != nil {
			fmt.Printf("Ошибка чтения параметра: %v\n", err)
			continue
		}

		fmt.Printf("\nТекущее значение параметра '%s':\n", paramName)
		fmt.Printf("Тип: %s\n", getTypeName(valueType))
		fmt.Printf("Значение: %v\n", currentValue)

		fmt.Print("\nВведите новое значение (Enter для отмены): ")
		var newValue string
		fmt.Scanln(&newValue)

		if newValue == "" {
			fmt.Println("Изменение отменено")
			continue
		}

		key, err = openRegistryKeyForWrite(regPath)
		if err != nil {
			fmt.Printf("Ошибка открытия раздела для записи: %v\n", err)
			continue
		}

		err = writeRegistryValue(key, paramName, newValue, valueType)
		key.Close()

		if err != nil {
			fmt.Printf("Ошибка при изменении значения: %v\n", err)
		} else {
			fmt.Println("Значение успешно изменено!")
		}

		fmt.Println(strings.Repeat("-", 50))
	}
}

// openRegistryKeyForEdit Открытие раздела для чтения
func openRegistryKeyForEdit(path string) (registry.Key, error) {
	parts := strings.SplitN(path, "\\", 2)
	rootKey := parts[0]
	subKey := ""
	if len(parts) > 1 {
		subKey = parts[1]
	}

	var root registry.Key
	switch strings.ToUpper(rootKey) {
	case "HKCR", "HKEY_CLASSES_ROOT":
		root = registry.CLASSES_ROOT
	case "HKCU", "HKEY_CURRENT_USER":
		root = registry.CURRENT_USER
	case "HKLM", "HKEY_LOCAL_MACHINE":
		root = registry.LOCAL_MACHINE
	case "HKU", "HKEY_USERS":
		root = registry.USERS
	case "HKCC", "HKEY_CURRENT_CONFIG":
		root = registry.CURRENT_CONFIG
	default:
		return 0, fmt.Errorf("неизвестный корневой ключ: %s", rootKey)
	}

	if subKey == "" {
		return root, nil
	}
	return registry.OpenKey(root, subKey, registry.QUERY_VALUE)
}

// openRegistryKeyForWrite Открытие раздела для записи
func openRegistryKeyForWrite(path string) (registry.Key, error) {
	parts := strings.SplitN(path, "\\", 2)
	rootKey := parts[0]
	subKey := ""
	if len(parts) > 1 {
		subKey = parts[1]
	}

	var root registry.Key
	switch strings.ToUpper(rootKey) {
	case "HKCU", "HKEY_CURRENT_USER":
		root = registry.CURRENT_USER
	case "HKLM", "HKEY_LOCAL_MACHINE":
		root = registry.LOCAL_MACHINE
	default:
		return 0, fmt.Errorf("для записи доступны только HKCU и HKLM (текущий: %s)", rootKey)
	}

	if subKey == "" {
		return root, nil
	}
	return registry.OpenKey(root, subKey, registry.SET_VALUE)
}

// readRegistryValue Чтение значения из реестра
func readRegistryValue(key registry.Key, name string) (interface{}, uint32, error) {
	// Пробуем разные типы
	if val, valType, err := key.GetStringValue(name); err == nil {
		return val, valType, nil
	}
	if val, valType, err := key.GetIntegerValue(name); err == nil {
		return val, valType, nil
	}
	if val, valType, err := key.GetBinaryValue(name); err == nil {
		return val, valType, nil
	}
	if val, valType, err := key.GetStringsValue(name); err == nil {
		return val, valType, nil
	}
	return nil, 0, fmt.Errorf("параметр не найден или имеет неподдерживаемый тип")
}

// writeRegistryValue Запись значения в реестр
func writeRegistryValue(key registry.Key, name, value string, valueType uint32) error {
	switch valueType {
	case registry.SZ, registry.EXPAND_SZ:
		return key.SetStringValue(name, value)
	case registry.DWORD, registry.QWORD:
		var intVal uint64
		fmt.Sscanf(value, "%d", &intVal)
		if valueType == registry.DWORD {
			return key.SetDWordValue(name, uint32(intVal))
		}
		return key.SetQWordValue(name, intVal)
	case registry.MULTI_SZ:
		stringList := strings.Split(value, ",")
		return key.SetStringsValue(name, stringList)
	default:
		return fmt.Errorf("неподдерживаемый тип для записи")
	}
}

// getTypeName Получение названия типа
func getTypeName(valueType uint32) string {
	switch valueType {
	case registry.SZ:
		return "REG_SZ"
	case registry.EXPAND_SZ:
		return "REG_EXPAND_SZ"
	case registry.BINARY:
		return "REG_BINARY"
	case registry.DWORD:
		return "REG_DWORD"
	case registry.QWORD:
		return "REG_QWORD"
	case registry.MULTI_SZ:
		return "REG_MULTI_SZ"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", valueType)
	}
}
