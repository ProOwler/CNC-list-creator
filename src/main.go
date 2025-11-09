package main

import (
	"encoding/xml"
	//	"errors"
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
	TargetDir      string         `xml:"TargetDir"`
	WorkReportFile string         `xml:"WorkReportFile"`
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
	dirSource  string   // Исходная папка для сканирования (из файла настроек)
	dirTarget  string   // Целевая папка
	fileReport string   // Файл отчета
}

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

// XML-представление отчёта
type XReportHead struct {
	XMLName        xml.Name        `xml:"Root"`
	ReportItemList XReportItemList `xml:"ReportItemList"`
}

type XReportItemList struct {
	ReportItem []XReportItem `xml:"ReportItem"`
}

type XReportItem struct {
	ItemName       string          `xml:"ItemName,attr"`
	Status         string          `xml:"Status,attr"`
	DateReady      string          `xml:"DateReady,attr"`
	Level          int             `xml:"Level,attr"`
	ReportItemList XReportItemList `xml:"ReportItemList,omitempty"`
}

/*
<Root>
	<ReportItemList>
		<ReportItem ItemName="Тест-заказ" Status="Ожидает" DateReady="" Level="1">
			<ReportItemList>
				<ReportItem ItemName="ЛДСП Белый Шагрень" Status="Готов" DateReady="2025-06-01" Level="0">
					<ReportItemList></ReportItemList>
				</ReportItem>
				<ReportItem ItemName="ЛДСП Белый ГП" Status="Ожидает" DateReady="" Level="0">
					<ReportItemList></ReportItemList>
				</ReportItem>
			</ReportItemList>
		</ReportItem>
	</ReportItemList>
</Root>
*/

// GO-представление отчёта
type ReportObj struct {
	itemName   string
	status     string
	dateReady  string
	level      int
	innerItems []ReportObj
}

/*
Шаблон отчёта
// массив объектов, в объекте массив проектов, в проекте массив раскроев
// на каждом уровне статус и дата выполнения
// статусы: ожидает / готов / иное
	[{
		"item_name" : string,
		"status" : string,
		"date_ready" : date,
		"level" : int,
		"inner_items" : [{ "item_name", "status", "date_ready", "level", "inner_items" }, ... ]
	}, ... ]
*/

// myMap: Пользовательский тип для хранения сопоставлений (например, кодов и расширений файлов)
type myMap map[string]string

// --- Глобальные переменные и константы ---

// имя файла настроек
const settingsFileName = "listMaker_settings.xml"

// Имя файла, генерируемого в каждой папке
const listFileName = "list.xml"

// стоп-слова, наличие которых надо проверять в именах файлов
var stopWords = []string{"fasady", "list", "ready"}

// статусы обработки папок
const (
	c_ST_OTHER   string = "Иное"
	c_ST_READY   string = "Готов"
	c_ST_PENDING string = "Ожидает"
)

// константы для разбития строки на части
const (
	c_PRT_DETAIL int = 2
	c_PRT_DATE   int = -1
)

// константы, как обрабатывать файлы в папке
const (
	c_PROC_NO = iota
	c_PROC_XML
	c_PROC_MPR
)

var listOfFileFormats = make(myMap)

// --- Основная функция ---

/**
 * main: Точка входа программы.
 * 1. Инициализирует настройки (IgnoreList и др.) из XML-файла.
 * 2. Если настройки не загружены, создает файл настроек по умолчанию и выходит.
 * 3. Определяет стартовую директорию: из аргумента командной строки или из настроек.
 * 4. Запускает обработку стартовой директории.
 * 5. Измеряет и выводит время выполнения.
 */
func main() {
	tThen := time.Now()
	//log.Println("Запуск программы...")
	//log.Println(os.Args[1])

	// 1. Загрузка настроек (нужны для IgnoreList и др.)
	settingsStruct, err := initSettings(settingsFileName)
	if err != nil {
		log.Printf("Ошибка чтения настроек (%s): %v. Создание файла настроек по умолчанию.\n", settingsFileName, err)
		checkFatal(writeDefaultSettingsToFile(settingsFileName), "Не удалось создать файл настроек по умолчанию\n")
		log.Printf("Файл настроек по умолчанию '%s' создан. Пожалуйста, отредактируйте его и перезапустите программу.\n", settingsFileName)
		// Выход, так как без базовых настроек (особенно IgnoreList) работа некорректна
		fmt.Scanln()
		return
	} else {
		log.Printf("Настройки успешно загружены из %s.", settingsFileName)
		log.Printf("Игнорируемые папки: %v", settingsStruct.ignoreList)
	}

	// 2. Определение стартовой директории
	var startDir string

	if len(os.Args) > 1 {
		progDir := filepath.Dir(os.Args[0]) // Директория, откуда запущена программа
		// Используем аргумент командной строки
		startDir = getAbsoluteFilepath(progDir, os.Args[1]) // Делаем путь абсолютным относительно папки программы
		//log.Printf("Используется стартовая папка из аргумента командной строки: %s", startDir)
	} else {
		// Используем папку из настроек
		startDir = settingsStruct.dirSource // Путь уже абсолютный после initSettings
		// log.Printf("Аргумент командной строки не найден. Используется стартовая папка из настроек: %s", startDir)
	}

	// Проверка, что startDir не пустая (на всякий случай)
	if startDir == "" {
		log.Println("Ошибка: Стартовая директория не определена (ни через аргумент, ни в настройках).")
		return
	}

	// 3. Запуск обработки
	processSourceDirectory(startDir, settingsStruct) // Передаем определенную startDir и настройки

	fmt.Printf("Стартовая папка фактическая: %s\n", startDir)
	fmt.Printf("Выполнение завершено. Затрачено времени: %.6f сек\n", time.Since(tThen).Seconds())
	fmt.Println("\nДля закрытия окна нажмите Enter")
	fmt.Scanln() // Раскомментируйте, если нужно оставлять консоль открытой после выполнения
}

// --- Функции обработки ---

/**
 * processSourceDirectory: Запускает рекурсивный обход и обработку указанной стартовой директории.
 * @param startDir - Абсолютный путь к директории, с которой начинается обработка.
 * @param settings - Загруженные настройки программы (для доступа к списку игнорирования).
 */
func processSourceDirectory(startDir string, settings InnerSettings) {
	log.Printf("Начало обработки папки: %s", startDir)

	// Определение форматов файлов для обработки (из второй программы)
	listOfFileFormats["7"] = "mpr"  // Код "7" для файлов .mpr
	listOfFileFormats["11"] = "xml" // Код "11" для файлов .xml

	// Запуск рекурсивного обхода из startDir
	sort.Strings(stopWords)
	reports := recursiveWalkthrough(startDir, settings).innerItems
	// Сохранение отчёта в файл
	validTimeName := strings.ReplaceAll(time.Now().Format(time.DateTime), ":", "-")
	reportFileFullName := filepath.Join(settings.dirTarget, strings.ReplaceAll(validTimeName, " ", "_")+"_"+settings.fileReport)
	createFile(reportFileFullName, []byte(createReport(reports)))
	// перемещение папок с готовыми заданиями, не работает при открытом окне проводника
	for _, proj := range reports {
		if proj.status == c_ST_READY {
			dateDirShort := proj.dateReady[0:7]
			dateDirFull := filepath.Join(settings.dirTarget, dateDirShort)
			if !isValidDir(dateDirFull) {
				os.MkdirAll(dateDirFull, 0777)
				if !isValidDir(dateDirFull) {
					log.Printf("Папка %s всё ещё недоступна", dateDirFull)
				}
			}
			err0 := os.Rename(
				filepath.Join(startDir, proj.itemName),
				filepath.Join(dateDirFull, proj.itemName))
			if err0 != nil {
				log.Printf("Ошибка перемещения директории %s: %v\n\nЗакройте окно Проводника!\n", proj.itemName, err0)
			}
		}
	}
}

/**
 * recursiveWalkthrough: Рекурсивно обходит директории, обрабатывает файлы и создает list.xml.
 * @param currentPath - Текущая директория для обхода.
 * @param settings - Настройки программы (для доступа к списку игнорирования).
 */
func recursiveWalkthrough(currentPath string, settings InnerSettings) ReportObj {
	// Получаем список содержимого текущей директории
	currentPathShort := filepath.Base(currentPath)
	dirEntries, err := os.ReadDir(currentPath)
	if err != nil {
		log.Printf("Ошибка чтения директории %s: %v", currentPath, err)
		return ReportObj{}
	}

	// алг - всё содержимое осматриваемой папки разделить на 2 перечня - [подпапки, файлы]
	var dirEntriesFileNames, dirEntriesDirNames, fullnamesToProceed []string
	for _, entry := range dirEntries {
		entryFullPath := filepath.Join(currentPath, entry.Name())
		if entry.IsDir() {
			// Проверяем, нужно ли игнорировать эту директорию
			if settings.isIgnored(entryFullPath) {
				continue // Пропускаем игнорируемую папку
			}
			dirEntriesDirNames = append(dirEntriesDirNames, entryFullPath)
		} else {
			dirEntriesFileNames = append(dirEntriesFileNames, entryFullPath)
		}
	}

	if len(dirEntriesFileNames) > 0 {
		sort.Strings(dirEntriesFileNames)
		// алг - если есть файл "плейлист" (list.xml),
		if hasStringInList(listFileName, dirEntriesFileNames) {
			//log.Println("Есть файл-список заданий")
			return ReportObj{
				itemName:  currentPathShort,
				level:     0,
				dateReady: "",
				status:    c_ST_PENDING,
			}
		}
		for _, fileName := range dirEntriesFileNames {
			if strings.Contains(filepath.Base(fileName), "ready") {
				// алг - если есть файл "плейлист фасадов" выполненный (ready_fasady.xml),
				if strings.Contains(filepath.Base(fileName), "fasady") {
					log.Printf("Путь: %s. Переместите файл ready_fasady.xml в папки с фасадами\n", currentPath)
					return ReportObj{
						itemName:  currentPathShort,
						level:     0,
						dateReady: "",
						status:    c_ST_PENDING,
					}
				}
				// алг - если есть файл-метка-отчёт order_ready_yyyymmdd.xml,
				if strings.Contains(filepath.Base(fileName), "order") {
					if dateString := getReadyDate(filepath.Base(fileName)); dateString != "" {
						innerObjects := getReportObjectsFromFile(fileName)
						lvl := 0
						for _, rep := range innerObjects {
							if rep.level >= lvl {
								lvl = rep.level + 1
							}
						}
						return ReportObj{
							itemName:   currentPathShort,
							level:      lvl,
							dateReady:  dateString,
							status:     c_ST_READY,
							innerItems: innerObjects,
						}
					} else {
						log.Printf("Ошибка извлечения даты из имени файла %s\n", fileName)
						return ReportObj{
							itemName:   currentPathShort,
							level:      0,
							dateReady:  dateString,
							status:     c_ST_PENDING,
							innerItems: getReportObjectsFromFile(fileName),
						}
					}
				}
				// алг - если есть выполненный файл "плейлист" (ready_yyyymmdd.xml),
				if dateString := getReadyDate(filepath.Base(fileName)); dateString != "" {
					return ReportObj{
						itemName:  currentPathShort,
						level:     0,
						dateReady: dateString,
						status:    c_ST_READY,
					}
				} else {
					log.Printf("Ошибка извлечения даты из имени файла %s\n", fileName)
					return ReportObj{
						itemName:  currentPathShort,
						level:     0,
						dateReady: dateString,
						status:    c_ST_PENDING,
					}
				}
			}
			// алг - если есть подходящие для обработки файлы-задания, обработать их,
			// пропускаем файлы со стоп-словами
			if hasStopWord(filepath.Base(fileName)) {
				continue
			}
			// MPR
			if strings.ToLower(getExtention(fileName)) == "mpr" {
				fullnamesToProceed = append(fullnamesToProceed, fileName)
			}
			// XML
			if strings.ToLower(getExtention(fileName)) == "xml" {
				updateFileWithXML(fileName)
				fullnamesToProceed = append(fullnamesToProceed, fileName)
			}
		}
		// создать плейлист
		if len(fullnamesToProceed) > 0 {
			outputXMLString := getOutputXML(fullnamesToProceed, listOfFileFormats)
			outputFilePath := filepath.Join(currentPath, listFileName)
			createFile(outputFilePath, []byte(outputXMLString))
			//	сформировать отчёт с записью о том, что папка в работе (статус ОЖИДАЕТ)
			//	ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
			return ReportObj{
				itemName:  currentPathShort,
				level:     0,
				dateReady: "",
				status:    c_ST_PENDING,
			}
		}
	}

	if len(dirEntriesDirNames) > 0 {
		sort.Strings(dirEntriesDirNames)
		var statuses, dates []string
		var childReports []ReportObj
		var lev int
		for _, dirName := range dirEntriesDirNames {
			child := recursiveWalkthrough(dirName, settings)
			st := child.status
			if child.level > lev {
				lev = child.level
			}
			if st == c_ST_OTHER {
				log.Printf("Требуется участие пользователя: статус %s у папки %s\n", st, dirName)
				return ReportObj{
					itemName:  currentPathShort,
					level:     lev + 1,
					dateReady: "",
					status:    c_ST_OTHER,
				}
			}
			statuses = append(statuses, st)
			dates = append(dates, child.dateReady)
			childReports = append(childReports, child)
		}
		if hasStringInList(c_ST_PENDING, statuses) {
			return ReportObj{
				itemName:   currentPathShort,
				level:      lev + 1,
				dateReady:  "",
				status:     c_ST_PENDING,
				innerItems: childReports,
			}
		} else {
			sort.Strings(dates)
			readyDate := dates[len(dates)-1]
			resReport := ReportObj{
				itemName:   currentPathShort,
				level:      lev + 1,
				dateReady:  readyDate,
				status:     c_ST_READY,
				innerItems: childReports,
			}
			fileShortName := "order_ready_" + readyDate[0:4] + readyDate[5:7] + readyDate[8:] + ".xml"
			resReport.writeReportToFile(filepath.Join(currentPath, fileShortName))
			return resReport
		}
	}
	return ReportObj{
		itemName:  currentPathShort,
		level:     0,
		dateReady: "",
		status:    c_ST_OTHER,
	}
}

func getReadyDate(shortFileName string) string {
	datePart := getPartFromDividedString(strings.TrimSuffix(shortFileName, filepath.Ext(shortFileName)), c_PRT_DATE)
	if len(datePart) != 8 {
		return ""
	} else {
		return datePart[0:4] + "-" + datePart[4:6] + "-" + datePart[6:]
	}
}

func getReportObjectsFromFile(fullFileName string) []ReportObj {
	myFileBytes, err := os.ReadFile(fullFileName)
	if err != nil {
		log.Printf("Не удалось прочитать файл отчёта %s: %w\n", fullFileName, err)
		return []ReportObj{{}}
	}
	var myRepXML XReportHead
	err = xml.Unmarshal(myFileBytes, &myRepXML)
	if err != nil {
		log.Printf("Не удалось разобрать XML из файла отчёта %s: %w\n", fullFileName, err)
		return []ReportObj{{}}
	}
	return getReportObjects(myRepXML)
}

func (item *ReportObj) writeReportToFile(fullFilePath string) {
	var objects []ReportObj
	objects = append(objects, *item)
	xmlReport := getReportXML(objects)
	myHeader := `<?xml version="1.0" encoding="utf-8" ?>` + "\n"
	xmlReportString := ""
	xmlReportBytes, errMarshal := xml.MarshalIndent(xmlReport, "", "	") // Используем табуляцию для отступов
	if errMarshal != nil {
		log.Printf("Ошибка при сериализации XML: %v", errMarshal)
		return
	}

	xmlReportString = myHeader + string(xmlReportBytes)
	createFile(fullFilePath, []byte(xmlReportString))
}

func getReportXML(itemObj []ReportObj) XReportHead {
	var result = XReportHead{}
	for _, entry := range itemObj {
		result.ReportItemList.ReportItem = append(result.ReportItemList.ReportItem, entry.convertReportItemToXML())
	}

	return result
}

func (item *ReportObj) convertReportItemToXML() XReportItem {
	var result = XReportItem{
		ItemName:  item.itemName,
		Level:     item.level,
		DateReady: item.dateReady,
		Status:    item.status,
	}
	for _, entry := range item.innerItems {
		result.ReportItemList.ReportItem = append(result.ReportItemList.ReportItem, entry.convertReportItemToXML())
	}
	return result
}

func getReportObjects(itemX XReportHead) []ReportObj {
	var result = []ReportObj{}
	for _, entry := range itemX.ReportItemList.ReportItem {
		result = append(result, entry.convertReportItemToObj())
	}
	return result
}

func (item *XReportItem) convertReportItemToObj() ReportObj {
	var result = ReportObj{
		itemName:  item.ItemName,
		level:     item.Level,
		dateReady: item.DateReady,
		status:    item.Status,
	}
	for _, entry := range item.ReportItemList.ReportItem {
		result.innerItems = append(result.innerItems, entry.convertReportItemToObj())
	}
	return result
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
	// Определяем абсолютный путь к файлу настроек относительно папки программы
	progDir := filepath.Dir(os.Args[0])
	absolutePath := getAbsoluteFilepath(progDir, pathToFileWithSettings)
	log.Printf("Попытка чтения файла настроек: %s", absolutePath)
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
		return fmt.Errorf("Не удалось прочитать файл настроек %s: %w\n", fileAbsolutePath, err)
	}

	var fileSettings XMLSettings
	err = xml.Unmarshal(myFileBytes, &fileSettings)
	if err != nil {
		return fmt.Errorf("Не удалось разобрать XML из файла настроек %s: %w\n", fileAbsolutePath, err)
	}

	// Заполнение внутренней структуры настроек
	settings.ignoreList = []string{}
	for _, el := range fileSettings.IgnoreDirList.IgnoreDir {
		settings.ignoreList = append(settings.ignoreList, el.Name)
	}
	settings.dirTarget = getAbsoluteFilepath(filepath.Dir(fileAbsolutePath), fileSettings.TargetDir)
	settings.fileReport = fileSettings.WorkReportFile // Храним только имя файла

	// Валидация настроек (Если SourceDir пуст, станет ".")
	if fileSettings.SourceDir == "" {
		settings.dirSource = getAbsoluteFilepath(filepath.Dir(fileAbsolutePath), ".")
	} else {
		settings.dirSource = getAbsoluteFilepath(filepath.Dir(fileAbsolutePath), fileSettings.SourceDir)
	}

	// Логируем прочитанные настройки
	log.Println("Настройки прочитаны из файла:")
	log.Printf("  SourceDir (из файла): %s", settings.dirSource)
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
func (settings *InnerSettings) isIgnored(dirPath string) bool {
	if isValidDir(dirPath) {
		// Получаем только имя папки из полного пути
		dirName := filepath.Base(dirPath)

		// Имя папки не может быть пустым или "." или ".."
		if dirName == "" || dirName == "." || dirName == ".." {
			return false
		}

		// Проверяет наличие папки в списке игнорирования
		return hasStringInList(dirName, settings.ignoreList)
	} else {
		return false
	}
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
		<IgnoreDir Name="#Archive"/>
		<IgnoreDir Name="#Frezerovki"/>
		<IgnoreDir Name="#Без_кромок"/>
		<IgnoreDir Name="#ВЫПОЛНЕННЫЕ"/>
		<IgnoreDir Name="#ЕВРОЗАПИЛ"/>
		<IgnoreDir Name="#КОММЕРЦИЯ"/>
		<IgnoreDir Name="1111"/>
		<IgnoreDir Name="123"/>
		<IgnoreDir Name="1234"/>
		<IgnoreDir Name="12345"/>
		<IgnoreDir Name=".git"/>
		<IgnoreDir Name=".svn"/>
	</IgnoreDirList>
	<SourceDir>.</SourceDir>
	<TargetDir>./#ВЫПОЛНЕННЫЕ</TargetDir>
	<WorkReportFile>WorkReport.txt</WorkReportFile>
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
	err := createFile(fileAbsolutePath, []byte(xmlString))
	return err
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

	//log.Printf("Обновление XML-файла: %s", filePath)
	myFileBytes, errRead := os.ReadFile(filePath)
	if errRead != nil {
		log.Printf("Ошибка чтения XML-файла %s для обновления: %v", filePath, errRead)
		return
	}

	// Получаем обновленное содержимое XML
	myEditedXML, xmlUpdated, errUpdate := getUpdatedXML(myFileBytes)
	if errUpdate != nil {
		// Ошибка уже залогирована внутри getUpdatedXML
		return
	}

	// Перезаписываем файл с обновленным содержимым
	if xmlUpdated {
		createFile(filePath, []byte(myEditedXML))
	}
}

/**
 * getUpdatedXML: Разбирает XML байты, обновляет поле Name у панелей и возвращает обновленный XML в виде строки.
 * @param inXMLBytes - Содержимое XML-файла в виде байтов.
 * @return string - Обновленное XML-содержимое в виде строки (с заголовком).
 * @return bool - true, если строка обновлена.
 * @return error - Ошибка при разборе или сериализации XML.
 */
func getUpdatedXML(inXMLBytes []byte) (string, bool, error) {
	var root XResult

	err := xml.Unmarshal(inXMLBytes, &root)
	if err != nil {
		log.Printf("Ошибка при разборе XML для обновления: %v", err)
		return "", false, err // Возвращаем ошибку
	}

	// Обновляем поле Name для каждой панели
	updated := false // Флаг, что хотя бы одно имя было обновлено
	for i := range root.Project.Panels.Panel {
		panel := &root.Project.Panels.Panel[i]
		width64, errW := strconv.ParseFloat(strings.Replace(panel.Width, ",", ".", 1), 64)
		length64, errL := strconv.ParseFloat(strings.Replace(panel.Length, ",", ".", 1), 64)
		thickness64, errT := strconv.ParseFloat(strings.Replace(panel.Thickness, ",", ".", 1), 64)

		if errW != nil || errL != nil || errT != nil {
			log.Printf("Предупреждение: Не удалось преобразовать Длину ('%s'), Ширину ('%s') или Толщину ('%s') в число для панели ID='%s'. Имя не будет обновлено.", panel.Length, panel.Width, panel.Thickness, panel.ID)
			continue // Пропускаем эту панель, если размеры некорректны
		}

		// Используем .0f, чтоб не было знаков после запятой
		newName := fmt.Sprintf("%.0f_%.0f_%.0f", length64, width64, thickness64)
		if panel.Name != newName {
			panel.Name = newName
			updated = true
		}
	}

	if !updated {
		//log.Println("Обновление XML не требуется, имена панелей уже соответствуют формату Длина_Ширина_Толщина.")
		// Возвращаем пустую строку, чтобы избежать лишней сериализации
		return "", false, nil
	}

	// Сериализуем обновленную структуру обратно в XML
	updatedXMLBytes, errMarshal := xml.MarshalIndent(root, "", "	") // Используем табуляцию для отступов
	if errMarshal != nil {
		log.Printf("Ошибка при сериализации обновленного XML: %v", errMarshal)
		return "", false, errMarshal // Возвращаем ошибку
	}

	myHeader := `<?xml version="1.0" encoding="utf-8" ?>` + "\n"
	updatedXML := ""
	updatedXML = myHeader + string(updatedXMLBytes)
	return updatedXML, true, nil
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
		sb.WriteString(getFiletypeCode(extCodes, getExtention(pathEntry))) // Получаем код типа файла
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
			log.Printf("Предупреждение: Не удалось извлечь количество деталей из имени файла '%s'. Запись в ProcessList не добавлена.", elemPath)
		}
	}
	return sb.String()
}

func createReport(reports []ReportObj) string {
	var reportStrings []string
	for _, rep := range reports {
		dateMonth := ""
		if rep.dateReady != "" {
			dateMonth = rep.dateReady[0:7]
		}
		reportStrings = append(reportStrings, dateMonth+" - "+rep.itemName+"\n")
	}
	sort.Strings(reportStrings)
	var sb strings.Builder
	for _, rs := range reportStrings {
		sb.WriteString(rs)
	}
	return sb.String()
}

// --- Вспомогательные функции

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

/** Возвращает код типа файла на основе его расширения
 * getFiletypeCode: Ищет значение в карте myMap и возвращает соответствующий ключ.
 * @param storage - Карта для поиска.
 * @param s - Значение для поиска (например, расширение файла).
 * @return string - Ключ (код) или пустая строка, если значение не найдено.
 */
func getFiletypeCode(storage myMap, s string) string {
	for k, v := range storage {
		if v == s {
			return k
		}
	}
	return ""
}

/**
 * hasStopWord: Проверяет наличие стоп-слов в строке (без учета регистра).
 * @param examinedStr - Проверяемая строка.
 * @return bool - true, если стоп-слово найдено, иначе false.
 */
func hasStopWord(examinedStr string) bool {
	for _, item := range stopWords {
		if strings.Contains(strings.ToLower(examinedStr), strings.ToLower(item)) {
			return true
		}
	}
	return false
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

/**
 * countDetails: Извлекает количество деталей из строки (кода детали).
 * Ожидает формат типа "КОД_КОЛИЧЕСТВО_..."
 * @param detailCode - Строка с кодом детали (обычно имя файла без расширения).
 * @return string - Строка с количеством или пустая строка, если не найдено или формат неверный.
 */
func countDetails(detailCode string) string {
	// Ожидаем как минимум 2 части (код_количество)
	// Если все проверки пройдены, возвращаем извлеченное количество
	return checkDetailsAmount(getPartFromDividedString(detailCode, c_PRT_DETAIL))
}

func getPartFromDividedString(filename string, flag int) string {
	parts := strings.Split(filename, "_")
	switch {
	case flag == c_PRT_DETAIL:
		if len(parts) < c_PRT_DETAIL {
			return ""
		} else {
			return parts[c_PRT_DETAIL-1]
		}
	case flag == c_PRT_DATE:
		if len(parts) == 0 {
			return ""
		} else {
			if resStr := parts[len(parts)-1]; len(resStr) != 8 {
				return ""
			} else {
				return resStr
			}
		}
	default:
		return ""
	}
}

/**
 * checkDetailsAmount: Проверяет строковое значение количества деталей.
 * @param inString - Строка с предположительно количеством деталей
 * @return string - Строка с количеством, если всё ОК, или пустая строка, если что-то пошло не так
 */
func checkDetailsAmount(inString string) string {
	if inString == "" {
		return ""
	}
	// Проверяем, что строка состоит только из цифр
	for _, r := range inString {
		if !unicode.IsDigit(r) {
			return "" // Если есть нецифровой символ, формат неверный
		}
	}
	// Если все проверки пройдены, возвращаем извлеченное количество
	return inString
}

/*
Общий алгоритм

первый запуск recursiveWalkthrough (далее "RWt()") не оставляет следов в осматриваемой папке,
	но во всех вложенных папках может оставить файл с меткой о выполнении заказа (метка - файл с названием order_ready_yyyymmdd.xml)
вместо этого первый запуск RWt() должен все выполненные заказы перемещать в целевую папку

в каждой папке, в которую RWt() зашла во время обхода, возможны взаимно исключающие исходы:
- создала метку о выполнении заказа // во всех подпапках есть файлы с подстрокой "ready" в названии
- создала файл list.xml и обработала (попыталась) XML-файлы // в этой папке (нет list.xml) И (есть подходящие файлы)
- ничего не произошло

Соответственно, // <- переработать
первым делом нужно всё содержимое осматриваемой папки разделить на 2 перечня - [подпапки, файлы],
далее, обработать перечень файлов:
	- если есть файл "плейлист" (list.xml),
		сформировать отчёт с записью о том, что папка в работе (статус ОЖИДАЕТ)
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
	- если есть файл "плейлист фасадов" выполненный (ready_fasady.xml),
		вывести сообщение "Путь: заказчик/заказ. Скопируйте файл ready_fasady.xml в папки с фасадами",
		сформировать отчёт с записью о том, что папка в работе (статус ОЖИДАЕТ)
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
	- если есть файл-метка-отчёт order_ready_yyyymmdd.xml,
		прочитать отчёт,
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
	- если есть выполненный файл "плейлист" (ready_yyyymmdd.xml),
		сформировать отчёт с записью о том, что папка в работе (статус ГОТОВ)
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
	- если есть подходящие для обработки файлы-задания,
		обработать их,
		сформировать отчёт с записью о том, что папка в работе (статус ОЖИДАЕТ)
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
	- иначе - переход к следующему этапу 🡗
после этого запустить во всех подпапках RWt()
	- если подпапок нет,
		сформировать отчёт с записью о том, что требуется внимание пользователя (статус ИНОЕ)
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
	- если хоть у одной папки вернулся статус ИНОЕ,
		сформировать отчёт с записью о том, что требуется внимание пользователя (статус ИНОЕ)
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
	- если во всех случаях вернулся статус ГОТОВ,
		сформировать отчёт с записью о том, что папка обработана (статус ГОТОВ)
		записать отчёт в файл (оставить метку),
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт
	- иначе (есть ОЖИДАЕТ и нет ИНОЕ)
		сформировать отчёт с записью о том, что папка в работе (статус ОЖИДАЕТ)
		ЗАВЕРШИТЬ выполнение функции, вернуть отчёт

кейсы по приоритетам и исходам:
	1) папка содержит list.xml
		=> вернуть отчёт с записью о том, что папка в работе (статус ОЖИДАЕТ)
	2) в папке есть файл ready_fasady.xml
		=> вывести сообщение "Путь: заказчик/заказ. Скопируйте файл ready_fasady.xml в папки с фасадами"
		=> вернуть отчёт с записью о том, что папка в работе (статус ОЖИДАЕТ)
	3) папка содержит метку готовности ИЛИ ready.xml
		=> сформировать и вернуть отчёт (статус ГОТОВ)
	4) папка содержит файлы с заданиями (xml или mpr)
		=> сформировать list.xml
		=> обработать файлы
		=> вернуть отчёт с записью о том, что папка в работе (статус ОЖИДАЕТ)
	5) во всех подпапках есть метка готовности,
		=> создать метку готовности
		=> сформировать и вернуть отчёт (статус ГОТОВ)
	6) НЕ во всех подпапках есть метка готовности,
		=> сформировать и вернуть отчёт (статус ОЖИДАЕТ или ИНОЕ, если есть ИНОЕ)
	7) ИНОЕ
		=> вернуть пустой отчёт (статус ИНОЕ)

Шаблон отчёта //
// массив объектов, в объекте массив проектов, в проекте массив раскроев
// на каждом уровне статус и дата выполнения
// статусы: ожидает / готов / иное
	[{
		"order_name" : string,
		"status" : string,
		"date_ready" : date,
		"projects" : [{
			"project_name" : string,
			"status" : string,
			"date_ready" : date,
			"cuts" : [{
				"material_name" : string,
				"status" : string,
				"date_ready" : date,
			}, ... ]
		}, ... ]
	}, ... ]

*/
