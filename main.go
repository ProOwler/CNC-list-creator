package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

/*

начиная с папки, переданной в качестве аргумента запуска,
	формировать список содержимого папки
	начиная с первого элемента в этой папке,
		если это файл,
			если его расширение входит в список,
				внести его полный путь в список
			иначе - пропустить
		иначе (папка)
			не вносить в список,
			запустить такую же функцию для этой папки
	сохранить список в файл с указанным названием
*/

type myMap map[string]string

// Обёртка для проверки ошибок, возвращаемых функциями
func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}

// Основная функция - запускает рекурсивный обход указанной папки и записывает результаты в файл с указанным (здесь же) названием
func main() {
	nameForListOfFiles := "list.xml"
	// nameForListOfFiles := "output.txt"

	listOfFileFormats := make(myMap)
	listOfFileFormats["7"] = "mpr"
	listOfFileFormats["11"] = "xml"

	recursiveWalkthrough(getStartDirPath(), nameForListOfFiles, listOfFileFormats)
}

// Рекурсивно обходит всё содержимое папки, которая была указана при запуске программы, создаёт файлы list.xml, если их нет
func recursiveWalkthrough(startPath string, outputFilename string, fileFormats myMap) {
	listOfDirContent := getListOfDirAndFiles(startPath)

	var resultList []string
	for i := 0; i < len(listOfDirContent); i++ {
		if checkIsDir(getAbsoluteFilepath(startPath, listOfDirContent[i])) {
			recursiveWalkthrough(
				getAbsoluteFilepath(startPath, listOfDirContent[i]+"\\"),
				outputFilename,
				fileFormats)
		} else {
			if getStringCode(fileFormats, strings.ToLower(getExtention(listOfDirContent[i]))) != "" {
				resultList = append(resultList, getAbsoluteFilepath(startPath, listOfDirContent[i]))
			}
		}
	}

	// TODO: Уточнить случаи, когда НЕ НАДО формировать итоговый файл
	mayProcessResultList := (len(resultList) > 0) && (!hasStringInList(outputFilename, listOfDirContent))

	if mayProcessResultList {
		myOutput := getOutputXML(resultList, fileFormats)
		myOutputFile, err1 := os.Create(filepath.Join(startPath, outputFilename))
		check(err1)

		_, err1 = myOutputFile.WriteString(myOutput)
		check(err1)

		check(myOutputFile.Close())
		//fmt.Println("Добавлен файл " + outputFilename)
	}
}

// Проверяет наличие строки в массиве строк
func hasStringInList(searchFor string, stringList []string) bool {
	sort.Strings(stringList)
	pos := sort.SearchStrings(stringList, searchFor)
	if pos >= len(stringList) {
		return false
	}
	res := searchFor == stringList[pos]
	return res
}

// Возвращает код строки, сохранённой в хранилище
func getStringCode(storage myMap, s string) string {
	res := ""
	for k, v := range storage {
		if v == s {
			return k
		}
	}
	return res
}

// Возвращает расширение файла
func getExtention(name string) string {
	result := filepath.Ext(name)
	if len(result) > 0 {
		return result[1:]
	} else {
		return ""
	}
}

// Возвращает путь к папке, содержимое которой нужно обработать
func getStartDirPath() string {
	return getDirOfArg(getMainStartupArg())
}

// Позволяет убедиться, что мы работаем с абсолютным путем к интересующему нас файлу или папке
func getAbsoluteFilepath(parent string, s string) string {
	if filepath.IsAbs(s) {
		return s
	} else {
		return filepath.Join(parent, s)
	}
}

// Возвращает путь к папке, чьи внутренности будут обрабатываться
func getMainStartupArg() string {
	if len(os.Args) > 1 {
		//fmt.Println("Скормлена папка: " + os.Args[1])
		return (os.Args[1])
	} else {
		//fmt.Println("Вызван сам по себе: \n" + os.Args[0])
		return os.Args[0]
	}
}

// Возвращает список содержимого переданной папки или пустой массив
func getListOfDirAndFiles(givenFilename string) []string {
	var myList []string

	if checkIsDir(givenFilename) {
		myFile, err1 := os.Open(givenFilename)
		check(err1)

		myList, err1 = myFile.Readdirnames(-1)
		check(err1)

		check(myFile.Close())
	}
	return myList
}

// Если дан путь к папке, его и возвращает, если к файлу, то возвращает путь к папке, в которой этот файл лежит
func getDirOfArg(givenFilename string) string {
	if checkIsDir(givenFilename) {
		return givenFilename
	} else {
		return filepath.Dir(givenFilename)
	}
}

// Проверяет, директория перед нами или нет
func checkIsDir(givenFilename string) bool {
	myFileInfo, err1 := os.Stat(givenFilename)
	check(err1)

	if myFileInfo.IsDir() {
		return true
	} else {
		return false
	}
}

// Выделяет в строке (имя файла) подстроку (количество экземпляров детали)
func countDetails(detailCode string) string {
	codeParts := strings.Split(detailCode, "_")
	if len(codeParts) < 3 {
		return ""
	}
	if codeParts[1] == "" {
		return ""
	}
	for _, elem := range codeParts[1] {
		if !unicode.IsDigit(elem) {
			return ""
		}
	}
	return codeParts[1]
}

// Формирует строку с итоговым XML
func getOutputXML(myList []string, extCodes myMap) string {
	resultString := "<WorkList><Version><Major>1</Major><Minor>0</Minor></Version><FileList>" +
		getXMLFileList(myList, extCodes) +
		"</FileList><ProcessList>" +
		getXMLProcessList(myList) +
		"</ProcessList></WorkList>"

	//	fmt.Println(resultString)

	return resultString
}

// Формирует часть итогового XML - список файлов
func getXMLFileList(myPathList []string, extCodes myMap) string {
	resString := ""
	for _, pathEntry := range myPathList {
		resString += "<Item><FileType>" +
			getFiletypeCode(pathEntry, extCodes) +
			"</FileType><FilePath>" +
			pathEntry +
			"</FilePath></Item>"
	}
	return resString
}

// Формирует часть итогового XML - список деталей для обработки с количеством экземпляров
func getXMLProcessList(myList []string) string {
	resString := ""
	detailCode := ""
	detailCount := ""
	for _, elem := range myList {
		detailCode = filepath.Base(elem)
		detailCount = countDetails(detailCode)
		if detailCount != "" {
			resString += "<Item><SerialNum>" +
				detailCode[:(len(detailCode)-4)] +
				"</SerialNum><PlanCount>" +
				detailCount +
				"</PlanCount><Count>0</Count></Item>"
		}
	}
	return resString
}

// Возвращает код типа файла - если вдруг используются задания не в XML
func getFiletypeCode(myPath string, extCodes myMap) string {
	res := ""
	for k, v := range extCodes {
		if strings.ToLower(getExtention(myPath)) == v {
			return k
		}
	}
	return res
}
