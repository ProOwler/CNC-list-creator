package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// --- Структуры и типы данных ---

// XMLSettings: Структура для чтения настроек из XML-файла
type XMLSettings struct {
	XMLName        xml.Name       `xml:"Root"`
	IgnoreDirList  XIgnoreDirList `xml:"IgnoreDirList"`
	SourceDir      string         `xml:"SourceDir"`
	TargetDir      string         `xml:"TargetDir"`      // Пока не используется в объединенной логике
	WorkReportFile string         `xml:"WorkReportFile"` // Пока не используется в объединенной логике
}

// XIgnoreDirList: Список игнорируемых директорий в XML
type XIgnoreDirList struct {
	IgnoreDir []XIgnoreDir `xml:"IgnoreDir"`
}

// XIgnoreDir: Игнорируемая директория в XML
type XIgnoreDir struct {
	Name string `xml:"Name,attr"`
}

// InnerSettings: Внутреннее представление настроек программы
type InnerSettings struct {
	ignoreList []string // Список имен папок, которые нужно игнорировать
	dirSource  string   // Исходная папка для сканирования
	dirTarget  string   // Целевая папка (пока не используется)
	fileReport string   // Файл отчета (пока не используется)
}

// myMap: Пользовательский тип для хранения сопоставлений (например, кодов и расширений файлов)
type myMap map[string]string

// XResult: Структура для разбора XML-файлов деталей (из второй программы)
type XResult struct {
	XMLName xml.Name `xml:"Root"`
	Project XProject `xml:"Project"`
}

// XProject: Структура проекта в XML детали
type XProject struct {
	Name   string  `xml:"Name,attr"`
	Flag   string  `xml:"Flag,attr"`
	Panels XPanels `xml:"Panels"`
}

// XPanels: Список панелей в XML детали
type XPanels struct {
	Panel []XPanel `xml:"Panel"`
}

// XPanel: Структура панели в XML детали
type XPanel struct {
	ID             string `xml:"ID,attr"`
	Name           string `xml:"Name,attr"` // Это поле будет обновлено
	Width          string `xml:"Width,attr"`
	Length         string `xml:"Length,attr"`
	Material       string `xml:"Material,attr"`
	Thickness      string `xml:"Thickness,attr"`
	IsProduce      string `xml:"IsProduce,attr"`
	MachiningPoint string `xml:"MachiningPoint,attr"`
	Type           string `xml:"Type,attr"`
	Face5ID        string `xml:"Face5ID,attr"`
	Face6ID        string `xml:"Face6ID,attr"`
	Grain          string `xml:"Grain,attr"`
	Count          string `xml:"Count,attr"`
	Machines       string `xml:",innerxml"`
	EdgeGroup      string `xml:",innerxml"`
}

// --- Глобальные переменные и константы ---
const settingsFileName = "report_settings.xml"
const listFileName = "list.xml" // Имя файла, генерируемого в каждой папке

// --- Основная функция ---

/**
 * main: Точка входа программы.
 * 1. Инициализирует настройки из XML-файла.
 * 2. Если настройки не загружены, создает файл настроек по умолчанию.
 * 3. Если настройки загружены, запускает обработку исходной директории.
 * 4. Измеряет и выводит время выполнения.
 */
func main() {
	tThen := time.Now()
	log.Println("Запуск программы...")

	settingsStruct, err := initSettings(settingsFileName)
	if err != nil {
		log.Printf("Ошибка чтения настроек (%s): %v. Создание файла настроек по умолчанию.", settingsFileName, err)
		// Используем Fatal, так как без настроек работа невозможна
		checkFatal(writeDefaultSettingsToFile(settingsFileName), "Не удалось создать файл настроек по умолчанию")
		log.Printf("Файл настроек по умолчанию '%s' создан. Пожалуйста, отредактируйте его и перезапустите программу.", settingsFileName)
	} else {
		log.Printf("Настройки успешно загружены из %s.", settingsFileName)
		log.Printf("Исходная папка: %s", settingsStruct.dirSource)
		log.Printf("Игнорируемые папки: %v", settingsStruct.ignoreList)
		processSourceDirectory(settingsStruct)
	}

	fmt.Printf("Выполнение завершено. Затрачено времени: %.6f сек\n", time.Since(tThen).Seconds())
	// fmt.Scanln() // Раскомментируйте, если нужно оставлять консоль открытой после выполнения
}

// --- Функции обработки ---

/**
 * processSourceDirectory: Запускает рекурсивный обход и обработку исходной директории.
 * @param settings - Загруженные настройки программы.
 */
func processSourceDirectory(settings InnerSettings) {
	log.Printf("Начало обработки директории: %s", settings.dirSource)

	// Определение форматов файлов для обработки (из второй программы)
	listOfFileFormats := make(myMap)
	listOfFileFormats["7"] = "mpr"  // Код "7" для файлов .mpr
	listOfFileFormats["11"] = "xml" // Код "11" для файлов .xml

	// Проверка существования исходной директории
	if _, err := os.Stat(settings.dirSource); os.IsNotExist(err) {
		log.Printf("Ошибка: Исходная директория '%s' не найдена.", settings.dirSource)
		return
	}

	recursiveWalkthrough(settings.dirSource, settings, listOfFileFormats)
	log.Println("Обработка директории завершена.")
}

/**
 * recursiveWalkthrough: Рекурсивно обходит директории, обрабатывает файлы и создает list.xml.
 * @param currentPath - Текущая директория для обхода.
 * @param settings - Настройки программы (для доступа к списку игнорирования).
 * @param fileFormats - Карта разрешенных форматов файлов и их кодов.
 */
func recursiveWalkthrough(currentPath string, settings InnerSettings, fileFormats myMap) {
	log.Printf("Сканирование папки: %s", currentPath)

	// Получаем список содержимого текущей директории
	dirEntries, err := os.ReadDir(currentPath)
	if err != nil {
		log.Printf("Ошибка чтения директории %s: %v", currentPath, err)
		return
	}

	var filesToProcess []string         // Список файлов для включения в list.xml в *этой* директории
	var currentDirContentNames []string // Имена файлов и папок в текущей директории

	for _, entry := range dirEntries {
		entryName := entry.Name()
		fullEntryPath := filepath.Join(currentPath, entryName)
		currentDirContentNames = append(currentDirContentNames, entryName)

		if entry.IsDir() {
			// Проверяем, нужно ли игнорировать эту директорию
			isIgnored, errIgnore := settings.isIgnored(fullEntryPath) // Проверяем по полному пути
			if errIgnore != nil {
				log.Printf("Ошибка проверки игнорирования для %s: %v", fullEntryPath, errIgnore)
				continue // Пропускаем папку, если не удалось проверить
			}
			if isIgnored {
				log.Printf("Игнорирование папки: %s", fullEntryPath)
				continue // Пропускаем игнорируемую папку
			}
			// Рекурсивный вызов для вложенной папки
			recursiveWalkthrough(fullEntryPath, settings, fileFormats)
		} else {
			// Это файл, проверяем расширение
			fileExt := strings.ToLower(getExtention(entryName))
			if getStringCode(fileFormats, fileExt) != "" {
				// Файл имеет одно из нужных расширений
				filesToProcess = append(filesToProcess, fullEntryPath)
				log.Printf("Найден файл для обработки: %s", fullEntryPath)
			}
		}
	}

	// Обработка найденных файлов и создание list.xml для *текущей* директории
	// Условия из второй программы:
	// 1) Есть файлы для обработки (filesToProcess не пуст)
	// 2) Файл list.xml еще не существует в этой папке
	shouldCreateListFile := (len(filesToProcess) > 0) && (!hasStringInSlice(listFileName, currentDirContentNames))

	if shouldCreateListFile {
		log.Printf("Создание %s в папке %s...", listFileName, currentPath)
		// Сначала обновляем XML файлы (если они есть в списке)
		for _, filePath := range filesToProcess {
			if strings.ToLower(getExtention(filePath)) == "xml" {
				updateFileWithXML(filePath) // Обновляет содержимое XML файла
			}
		}

		// Затем генерируем содержимое list.xml
		outputXMLString := getOutputXML(filesToProcess, fileFormats)
		outputFilePath := filepath.Join(currentPath, listFileName)

		// Записываем list.xml
		errWrite := os.WriteFile(outputFilePath, []byte(outputXMLString), 0644)
		if errWrite != nil {
			log.Printf("Ошибка записи файла %s: %v", outputFilePath, errWrite)
		} else {
			log.Printf("Файл %s успешно создан.", outputFilePath)
		}
	} else if len(filesToProcess) > 0 {
		log.Printf("Файл %s уже существует в папке %s, новый не создается.", listFileName, currentPath)
	}
}

// --- Функции работы с настройками (из первой программы) ---

/**
 * initSettings: Читает настройки из указанного файла.
 * @param pathToFileWithSettings - Путь к файлу настроек.
 * @return InnerSettings - Структура с настройками.
 * @return error - Ошибка, если чтение не удалось.
 */
func initSettings(pathToFileWithSettings string) (InnerSettings, error) {
	settingsStruct := InnerSettings{}
	absolutePath := getAbsoluteFilepath(filepath.Dir(os.Args[0]), pathToFileWithSettings)
	err := settingsStruct.readFromFile(absolutePath)
	return settingsStruct, err
}

/**
 * readFromFile: Метод для чтения настроек из XML-файла и заполнения структуры InnerSettings.
 * @receiver settings - Указатель на структуру InnerSettings для заполнения.
 * @param fileAbsolutePath - Абсолютный путь к файлу настроек.
 * @return error - Ошибка при чтении или разборе файла.
 */
func (settings *InnerSettings) readFromFile(fileAbsolutePath string) error {
	myFileBytes, err := os.ReadFile(fileAbsolutePath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл настроек %s: %w", fileAbsolutePath, err)
	}

	var fileSettings XMLSettings
	err = xml.Unmarshal(myFileBytes, &fileSettings)
	if err != nil {
		return fmt.Errorf("не удалось разобрать XML из файла настроек %s: %w", fileAbsolutePath, err)
	}

	// Заполнение внутренней структуры настроек
	settings.ignoreList = []string{}
	for _, el := range fileSettings.IgnoreDirList.IgnoreDir {
		settings.ignoreList = append(settings.ignoreList, el.Name)
	}
	// Сохраняем пути как есть (могут быть относительными или абсолютными)
	settings.dirSource = fileSettings.SourceDir
	settings.dirTarget = fileSettings.TargetDir
	settings.fileReport = fileSettings.WorkReportFile

	// Валидация настроек (проверяем, что SourceDir указан)
	if settings.dirSource == "" {
		return errors.New("в файле настроек не указан обязательный параметр <SourceDir>")
	}
	// Преобразуем SourceDir в абсолютный путь относительно файла настроек, если он не абсолютный
	settings.dirSource = getAbsoluteFilepath(filepath.Dir(fileAbsolutePath), settings.dirSource)

	log.Println("Настройки прочитаны:")
	log.Printf("  SourceDir: %s", settings.dirSource)
	log.Printf("  TargetDir: %s", settings.dirTarget)
	log.Printf("  WorkReportFile: %s", settings.fileReport)
	log.Printf("  IgnoreDirList: %v", settings.ignoreList)

	return nil
}

/**
 * isIgnored: Проверяет, соответствует ли имя директории одному из шаблонов в списке игнорирования.
 * Сравнивает *имя* папки, а не полный путь.
 * @receiver settings - Указатель на структуру InnerSettings.
 * @param dirPath - Полный путь к проверяемой директории.
 * @return bool - true, если директорию следует игнорировать.
 * @return error - Ошибка, если путь некорректен или не является директорией.
 */
func (settings *InnerSettings) isIgnored(dirPath string) (bool, error) {
	// Проверяем, что это действительно папка
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		// Если ошибка связана с тем, что файл/папка не найден, это не ошибка для этой функции
		if os.IsNotExist(err) {
			return false, nil // Не существующий путь не может быть игнорируемым
		}
		return false, fmt.Errorf("не удалось получить информацию о %s: %w", dirPath, err) // Другая ошибка Stat
	}
	if !fileInfo.IsDir() {
		// Это не папка, значит не игнорируем (игнорируем только папки)
		return false, nil
	}

	// Получаем только имя папки из полного пути
	dirName := filepath.Base(dirPath)

	// Имя папки не может быть пустым или "." или ".."
	if dirName == "" || dirName == "." || dirName == ".." {
		return false, nil
	}

	// Приводим список игнорирования к нижнему регистру для сравнения без учета регистра
	ignoreListLower := make([]string, len(settings.ignoreList))
	for i, ignoredName := range settings.ignoreList {
		ignoreListLower[i] = strings.ToLower(ignoredName)
	}
	// Сортируем для бинарного поиска
	sort.Strings(ignoreListLower)

	// Приводим имя текущей папки к нижнему регистру
	dirNameLower := strings.ToLower(dirName)

	// Ищем имя папки в отсортированном списке игнорирования
	pos := sort.SearchStrings(ignoreListLower, dirNameLower)

	// Проверяем, найдено ли точное совпадение
	if pos < len(ignoreListLower) && ignoreListLower[pos] == dirNameLower {
		return true, nil // Найдено в списке игнорирования
	}

	return false, nil // Не найдено
}

/**
 * writeDefaultSettingsToFile: Записывает XML-файл с настройками по умолчанию.
 * @param fileAbsolutePath - Абсолютный путь к файлу для записи.
 * @return error - Ошибка при записи файла.
 */
func writeDefaultSettingsToFile(fileAbsolutePath string) error {
	// Шаблон настроек по умолчанию (из первой программы)
	xmlString := `<?xml version="1.0" encoding="utf-8" ?>
<Root>
	<IgnoreDirList>
		<IgnoreDir Name="#ВЫПОЛНЕННЫЕ"/>
		<IgnoreDir Name="#Frezerovki"/>
		<IgnoreDir Name="#Archive"/>
		<IgnoreDir Name=".git"/>
		<IgnoreDir Name=".svn"/>
	</IgnoreDirList>
	<SourceDir>./input_files</SourceDir>
	<TargetDir>./processed_files</TargetDir>
	<WorkReportFile>WorkReport.xml</WorkReportFile>
</Root>`

	// Создаем директорию для файла настроек, если она не существует
	parentDir := filepath.Dir(fileAbsolutePath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		errMkdir := os.MkdirAll(parentDir, 0755)
		if errMkdir != nil {
			return fmt.Errorf("не удалось создать директорию %s: %w", parentDir, errMkdir)
		}
	}

	// Записываем файл
	err := os.WriteFile(fileAbsolutePath, []byte(xmlString), 0644)
	if err != nil {
		return fmt.Errorf("не удалось записать файл настроек %s: %w", fileAbsolutePath, err)
	}
	return nil
}

// --- Функции обработки файлов и XML (из второй программы) ---

/**
 * updateFileWithXML: Читает XML-файл, обновляет поле Name у панелей и перезаписывает файл.
 * @param filePath - Путь к XML-файлу для обновления.
 */
func updateFileWithXML(filePath string) {
	// Дополнительная проверка, что это XML (хотя вызывается только для XML)
	if strings.ToLower(getExtention(filePath)) != "xml" {
		return
	}

	log.Printf("Обновление XML-файла: %s", filePath)
	myFileBytes, errRead := os.ReadFile(filePath)
	if errRead != nil {
		log.Printf("Ошибка чтения XML-файла %s для обновления: %v", filePath, errRead)
		return
	}

	// Получаем обновленное содержимое XML
	myEditedXML, errUpdate := getUpdatedXML(myFileBytes)
	if errUpdate != nil {
		// Ошибка уже залогирована внутри getUpdatedXML
		return
	}

	// Перезаписываем файл с обновленным содержимым
	// Используем OpenFile + Truncate для перезаписи существующего файла
	myOutputFile, errOpen := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if errOpen != nil {
		log.Printf("Ошибка открытия XML-файла %s для записи: %v", filePath, errOpen)
		return
	}
	defer myOutputFile.Close() // Гарантируем закрытие файла

	_, errWrite := myOutputFile.WriteString(myEditedXML)
	if errWrite != nil {
		log.Printf("Ошибка записи обновленного XML в файл %s: %v", filePath, errWrite)
	} else {
		log.Printf("XML-файл %s успешно обновлен.", filePath)
	}
}

/**
 * getUpdatedXML: Разбирает XML байты, обновляет поле Name у панелей и возвращает обновленный XML в виде строки.
 * @param inXMLBytes - Содержимое XML-файла в виде байтов.
 * @return string - Обновленное XML-содержимое в виде строки (с заголовком).
 * @return error - Ошибка при разборе или сериализации XML.
 */
func getUpdatedXML(inXMLBytes []byte) (string, error) {
	var root XResult
	myHeader := `<?xml version="1.0" encoding="utf-8" ?>` + "\n"
	updatedXML := ""

	err := xml.Unmarshal(inXMLBytes, &root)
	if err != nil {
		log.Printf("Ошибка при разборе XML для обновления: %v", err)
		return "", err // Возвращаем ошибку
	}

	// Обновляем поле Name для каждой панели
	updated := false // Флаг, что хотя бы одно имя было обновлено
	for i := range root.Project.Panels.Panel {
		panel := &root.Project.Panels.Panel[i]
		// Используем .1f для одного знака после запятой
		width64, errW := strconv.ParseFloat(strings.Replace(panel.Width, ",", ".", 1), 64)
		length64, errL := strconv.ParseFloat(strings.Replace(panel.Length, ",", ".", 1), 64)

		if errW != nil || errL != nil {
			log.Printf("Предупреждение: Не удалось преобразовать Width ('%s') или Length ('%s') в число для панели ID='%s', Name='%s'. Имя не будет обновлено.", panel.Width, panel.Length, panel.ID, panel.Name)
			continue // Пропускаем эту панель, если размеры некорректны
		}

		newName := fmt.Sprintf("%.1f_%.1f", length64, width64)
		if panel.Name != newName {
			panel.Name = newName
			updated = true
		}
	}

	if !updated {
		log.Println("Обновление XML не требуется, имена панелей уже соответствуют формату Длина_Ширина.")
		// Возвращаем исходные байты с заголовком, чтобы избежать лишней сериализации
		return myHeader + string(inXMLBytes), nil
	}

	// Сериализуем обновленную структуру обратно в XML
	updatedXMLBytes, errMarshal := xml.MarshalIndent(root, "", "	") // Используем табуляцию для отступов
	if errMarshal != nil {
		log.Printf("Ошибка при сериализации обновленного XML: %v", errMarshal)
		return "", errMarshal // Возвращаем ошибку
	}

	updatedXML = myHeader + string(updatedXMLBytes)
	return updatedXML, nil
}

/**
 * getOutputXML: Формирует строку с итоговым XML для файла list.xml.
 * @param myPathList - Список полных путей к обработанным файлам (.mpr, .xml).
 * @param extCodes - Карта кодов для расширений файлов.
 * @return string - Строка с содержимым list.xml.
 */
func getOutputXML(myPathList []string, extCodes myMap) string {
	// Используем strings.Builder для эффективного построения строки
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="utf-8" ?>`)                // Добавляем заголовок XML
	sb.WriteString("\n<WorkList>\n")                                         // Открываем корневой элемент
	sb.WriteString("	<Version><Major>1</Major><Minor>0</Minor></Version>\n") // Версия
	sb.WriteString("	<FileList>\n")                                          // Секция списка файлов
	sb.WriteString(getXMLFileList(myPathList, extCodes))                     // Генерируем элементы Item для файлов
	sb.WriteString("	</FileList>\n")                                         // Закрываем секцию списка файлов
	sb.WriteString("	<ProcessList>\n")                                       // Секция списка процессов
	sb.WriteString(getXMLProcessList(myPathList))                            // Генерируем элементы Item для процессов
	sb.WriteString("	</ProcessList>\n")                                      // Закрываем секцию списка процессов
	sb.WriteString("</WorkList>\n")                                          // Закрываем корневой элемент

	return sb.String()
}

/**
 * getXMLFileList: Формирует часть XML (<Item>...</Item>) для списка файлов в list.xml.
 * @param myPathList - Список полных путей к файлам.
 * @param extCodes - Карта кодов для расширений.
 * @return string - XML-строка со списком файлов.
 */
func getXMLFileList(myPathList []string, extCodes myMap) string {
	var sb strings.Builder
	for _, pathEntry := range myPathList {
		sb.WriteString("		<Item>\n")
		sb.WriteString("			<FileType>")
		sb.WriteString(getFiletypeCode(pathEntry, extCodes)) // Получаем код типа файла
		sb.WriteString("</FileType>\n")
		sb.WriteString("			<FilePath>")
		// Экранируем специальные символы XML в пути к файлу
		xml.EscapeText(&sb, []byte(pathEntry))
		sb.WriteString("</FilePath>\n")
		sb.WriteString("		</Item>\n")
	}
	return sb.String()
}

/**
 * getXMLProcessList: Формирует часть XML (<Item>...</Item>) для списка процессов в list.xml.
 * Извлекает код детали и количество из имени файла.
 * @param myPathList - Список полных путей к файлам.
 * @return string - XML-строка со списком процессов.
 */
func getXMLProcessList(myPathList []string) string {
	var sb strings.Builder
	for _, elemPath := range myPathList {
		detailCodeWithExt := filepath.Base(elemPath)                                         // Получаем имя файла с расширением
		detailCode := strings.TrimSuffix(detailCodeWithExt, filepath.Ext(detailCodeWithExt)) // Убираем расширение
		detailCount := countDetails(detailCode)                                              // Извлекаем количество из имени файла

		if detailCount != "" { // Добавляем только если удалось извлечь количество
			sb.WriteString("		<Item>\n")
			sb.WriteString("			<SerialNum>")
			xml.EscapeText(&sb, []byte(detailCode)) // Экранируем код детали
			sb.WriteString("</SerialNum>\n")
			sb.WriteString("			<PlanCount>")
			xml.EscapeText(&sb, []byte(detailCount)) // Экранируем количество
			sb.WriteString("</PlanCount>\n")
			sb.WriteString("			<Count>0</Count>\n") // Поле Count по умолчанию 0
			sb.WriteString("		</Item>\n")
		} else {
			log.Printf("Предупреждение: Не удалось извлечь количество деталей из имени файла '%s'. Запись в ProcessList не добавлена.", detailCodeWithExt)
		}
	}
	return sb.String()
}

// --- Вспомогательные функции (объединенные и из обеих программ) ---

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

/**
 * getStringCode: Ищет значение в карте myMap и возвращает соответствующий ключ.
 * @param storage - Карта для поиска.
 * @param s - Значение для поиска (например, расширение файла).
 * @return string - Ключ (код) или пустая строка, если значение не найдено.
 */
func getStringCode(storage myMap, s string) string {
	for k, v := range storage {
		if v == s {
			return k
		}
	}
	return ""
}

/**
 * getFiletypeCode: Возвращает код типа файла на основе его расширения.
 * @param myPath - Путь к файлу.
 * @param extCodes - Карта кодов и расширений.
 * @return string - Код типа файла или пустая строка.
 */
func getFiletypeCode(myPath string, extCodes myMap) string {
	return getStringCode(extCodes, getExtention(myPath))
}

/**
 * hasStringInSlice: Проверяет наличие строки в срезе строк (без учета регистра).
 * @param searchFor - Строка для поиска.
 * @param stringSlice - Срез строк, в котором ищем.
 * @return bool - true, если строка найдена, иначе false.
 */
func hasStringInSlice(searchFor string, stringSlice []string) bool {
	searchLower := strings.ToLower(searchFor)
	for _, item := range stringSlice {
		if strings.ToLower(item) == searchLower {
			return true
		}
	}
	return false
}

/**
 * countDetails: Извлекает количество деталей из строки (кода детали).
 * Ожидает формат типа "КОД_КОЛИЧЕСТВО_..."
 * @param detailCode - Строка с кодом детали (обычно имя файла без расширения).
 * @return string - Строка с количеством или пустая строка, если не найдено или формат неверный.
 */
func countDetails(detailCode string) string {
	codeParts := strings.Split(detailCode, "_")
	// Ожидаем как минимум 2 части (код_количество)
	if len(codeParts) < 2 {
		return ""
	}
	// Вторая часть должна быть количеством
	countPart := codeParts[1]
	if countPart == "" {
		return ""
	}
	// Проверяем, что вторая часть состоит только из цифр
	for _, r := range countPart {
		if !unicode.IsDigit(r) {
			return "" // Если есть нецифровой символ, формат неверный
		}
	}
	// Если все проверки пройдены, возвращаем извлеченное количество
	return countPart
}
