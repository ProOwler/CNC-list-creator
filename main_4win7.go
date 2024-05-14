package main

import (
	//"slices"
	//"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	a = iota + 1
	_
	b
	c
	d = c + 2
	t
	i
	i2 = iota + 2 + 0.2
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

func main() {
	nameForListOfFiles := "list.xml"
	// nameForListOfFiles := "output.txt"

	listOfFileFormats := []string{"xml", "XML", "mpr", "MPR"}

	recursiveWalkthrough(getStartDirPath(), nameForListOfFiles, listOfFileFormats)

}

func recursiveWalkthrough(startPath string, outputFilename string, fileFormats []string) {
	listOfDirContent := getListOfDirAndFiles(startPath)

	var resultList []string
	for i := 0; i < len(listOfDirContent); i++ {
		if checkIsDir(getAbsoluteFilepath(startPath, listOfDirContent[i])) {
			recursiveWalkthrough(
				getAbsoluteFilepath(startPath, listOfDirContent[i]+"\\"),
				outputFilename,
				fileFormats)
		} else {
			if arrayContainsString(fileFormats, getExtention(listOfDirContent[i])) {
				resultList = append(resultList, getAbsoluteFilepath(startPath, listOfDirContent[i]))
			}
		}
	}

	if len(resultList) > 0 {
		myOutput := getOutputXML(resultList)
		myOutputFile := filepath.Join(startPath, outputFilename)
		check(os.WriteFile(myOutputFile, []byte(myOutput), 0666))
	}
}

func arrayContainsString(ar []string, s string) bool {
	res := false
	for _, el := range ar {
		if el == s {
			return true
		}
	}
	return res
}

func getExtention(name string) string {
	result := filepath.Ext(name)
	return result[1:]
}

func getStartDirPath() string {
	return getDirOfArg(getMainStartupArg())
}

func getAbsoluteFilepath(parent string, s string) string {
	if filepath.IsAbs(s) {
		return s
	} else {
		return filepath.Join(parent, s)
	}
}

func getMainStartupArg() string {
	if len(os.Args) > 1 {
		//fmt.Println("Скормлена папка: " + os.Args[1])
		return (os.Args[1])
	} else {
		//fmt.Println("Вызван сам по себе: \n" + os.Args[0])
		return os.Args[0]
	}
}

/*
func listOfStringsToOneString(stringsArg []string) string {
		resString := ""

		for i := 0; i < len(stringsArg); i++ {
			resString += stringsArg[i] + "\n"
		}
		return resString
	return strings.Join(stringsArg, "\n")
}
*/

func getListOfDirAndFiles(givenFilename string) []string {
	var myList []string

	if checkIsDir(givenFilename) {
		myFile, err1 := os.Open(givenFilename)
		check(err1)

		myList, err1 = myFile.Readdirnames(-1)
		check(err1)
	}
	return myList
}

func getDirOfArg(givenFilename string) string {
	if checkIsDir(givenFilename) {
		return givenFilename
	} else {
		return filepath.Dir(givenFilename)
	}
}

func checkIsDir(givenFilename string) bool {
	myFileInfo, err1 := os.Stat(givenFilename)
	check(err1)

	if myFileInfo.IsDir() {
		return true
	} else {
		return false
	}
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}

func countDetails(detailCode string) string {
	codeParts := strings.Split(detailCode, "_")
	if len(codeParts) < 3 {
		return ""
	}
	if codeParts[2] == "" {
		return ""
	}
	for _, elem := range codeParts[2] {
		if !unicode.IsDigit(elem) {
			return ""
		}
	}
	return codeParts[2]
}

func getOutputXML(myList []string) string {
	resultString := "<WorkList><Version><Major>1</Major><Minor>0</Minor></Version><FileList>" +
		getXMLFileList(myList) +
		"</FileList><ProcessList>" +
		getXMLProcessList(myList) +
		"</ProcessList></WorkList>"

	//	fmt.Println(resultString)

	return resultString
}

func getXMLFileList(myList []string) string {
	resString := "<Item><FileType>7</FileType><FilePath>"
	resString += strings.Join(myList, "</FilePath></Item><Item><FileType>7</FileType><FilePath>")
	resString += "</FilePath></Item>"
	return resString
}

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
