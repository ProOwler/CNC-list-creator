package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

/**
 * checkFatal: Проверяет ошибку и завершает программу с фатальной ошибкой, если она есть.
 * @param e - Проверяемая ошибка.
 * @param message - Сообщение для вывода перед завершением.
 */
func checkFatal(e error, message string) {
	if e != nil {
		log.Fatalf("%s: %v", message, e)
	}
}

func createFile(fullFilePath string, data []byte) error {
	errWrite := os.WriteFile(fullFilePath, data, 0644)
	if errWrite != nil {
		log.Printf("Ошибка записи файла %s: %v", fullFilePath, errWrite)
	}
	return errWrite
}

/**
 * getAbsoluteFilepath: Преобразует относительный путь в абсолютный, используя указанную родительскую директорию.
 * Если путь уже абсолютный, возвращает его без изменений.
 * @param parent - Родительская директория (абсолютный путь).
 * @param s - Путь для преобразования (может быть относительным или абсолютным).
 * @return string - Абсолютный путь.
 */
func getAbsoluteFilepath(parent string, s string) string {
	if filepath.IsAbs(s) {
		return filepath.Clean(s) // Возвращаем очищенный абсолютный путь
	}
	// Объединяем родительский путь и относительный путь, затем очищаем
	return filepath.Clean(filepath.Join(parent, s))
}

/**
 * getExtention: Возвращает расширение файла в нижнем регистре без точки.
 * @param name - Имя файла.
 * @return string - Расширение файла или пустая строка, если расширения нет.
 */
func getExtention(name string) string {
	ext := filepath.Ext(name)
	if len(ext) > 1 {
		return strings.ToLower(ext[1:]) // Убираем точку и приводим к нижнему регистру
	}
	return "" // Пустая строка, если нет расширения
}

// Проверяет наличие строки в массиве строк
func hasStringInList(searchFor string, stringList []string) bool {
	// Приводим массив к нижнему регистру для сравнения без учета регистра
	stringListLower := make([]string, len(stringList))
	for i, en := range stringList {
		stringListLower[i] = strings.ToLower(en)
	}
	// Приводим искомую строку к нижнему регистру
	searchForLower := strings.ToLower(searchFor)

	sort.Strings(stringListLower)
	pos := sort.SearchStrings(stringListLower, searchForLower)
	if pos >= len(stringListLower) {
		return false
	}
	res := searchForLower == stringListLower[pos]
	return res
}

func isValidDir(dirPath string) bool {
	// Проверяем, что это действительно папка
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		// Если ошибка связана с тем, что файл/папка не найден, это не ошибка для этой функции
		if os.IsNotExist(err) {
			log.Printf("Папка %s не существует: %v", dirPath, err)
			return false // Не существующий путь не может быть пригодным для использования
		}
		log.Printf("Не удалось получить информацию о %s: %v", dirPath, err) // Другая ошибка Stat
		return false
	}
	if !fileInfo.IsDir() {
		// Это не папка
		return false
	}
	return true
}
