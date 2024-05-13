package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
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
	// nameForListOfFiles := "list.xml"
	nameForListOfFiles := "output.txt"

	// чтоб компилятор не ругался на неиспользование массива
	listOfFileFormats := []string{"xml", "XML", "mpr", "MPR", "txt"}
	if len(listOfFileFormats) > 1 {
	}

	recursiveWalkthrough(getStartDirPath(), nameForListOfFiles, listOfFileFormats)

}

func recursiveWalkthrough(startPath string, outputFilename string, fileFormats []string) {
	fmt.Println("recursiveWalkthrough: " + startPath + " // " + outputFilename + " // " + listOfStringsToOneString(fileFormats))
	listOfDirContent := getListOfDirAndFiles(startPath)

	var resultList []string
	for i := 0; i < len(listOfDirContent); i++ {
		fmt.Println(getAbsoluteFilepath(listOfDirContent[i]))
		if checkIsDir(getAbsoluteFilepath(listOfDirContent[i])) {
			recursiveWalkthrough(
				getAbsoluteFilepath(listOfDirContent[i]+"\\"),
				outputFilename,
				fileFormats)
		} else {
			if slices.Contains(fileFormats, getExtention(listOfDirContent[i])) {
				resultList = append(resultList, getAbsoluteFilepath(listOfDirContent[i]))
			}
		}
	}

	myOutput := listOfStringsToOneString(resultList)
	myOutputFile := filepath.Join(startPath, outputFilename)
	//check(os.WriteFile(myOutputFile, []byte(myOutput), 0666))
	fmt.Println(myOutputFile + " / " + myOutput + "\n")
}

func getExtention(name string) string {
	fmt.Println("getExtention(" + name + ")\n")
	result := filepath.Ext(name)
	fmt.Println("Ext: " + result + "\n")
	return result[1:]
}

/*
func getOutput() string {
	return listOfStringsToOneString(getListOfDirAndFiles(getStartDirPath()))
}
*/

func getStartDirPath() string {
	fmt.Println("getStartDirPath()")
	return getDirOfArg(getAbsoluteFilepath(getMainStartupArg()))
}

func getAbsoluteFilepath(s string) string {
	fmt.Println("getAbsoluteFilepath(" + s + ")\n")
	absoluteFilepath, err := filepath.Abs(s)
	check(err)
	return absoluteFilepath
}

func getMainStartupArg() string {
	if len(os.Args) > 1 {
		fmt.Println("Скормлена папка: " + os.Args[1])
		return (os.Args[1])
	} else {
		fmt.Println("Вызван сам по себе, тестовая папка: \n" + "D:\\Prowler\\projects\\CNC-list-creator\\proj\\test")
		// return os.Args[0]
		return "D:\\Prowler\\projects\\CNC-list-creator\\proj\\test"
	}
}

func listOfStringsToOneString(strings []string) string {
	fmt.Println("listOfStringsToOneString(" + strings + ")\n\n")
	resString := ""

	for i := 0; i < len(strings); i++ {
		resString += strings[i] + "\n"
	}
	return resString
}

func getListOfDirAndFiles(givenFilename string) []string {
	fmt.Println("getListOfDirAndFiles(" + givenFilename + ")\n")
	var myList []string

	if checkIsDir(getAbsoluteFilepath(givenFilename)) {

		fmt.Println("--- " + getAbsoluteFilepath(givenFilename))
		myFile, err1 := os.Open(getAbsoluteFilepath(givenFilename))
		check(err1)

		myList, err1 = myFile.Readdirnames(-1)
		check(err1)
	}

	return myList
}

func getDirOfArg(givenFilename string) string {
	fmt.Println("getDirOfArg(" + givenFilename + ")\n")
	if checkIsDir(givenFilename) {
		return givenFilename
	} else {
		return filepath.Dir(givenFilename)
	}
}

func checkIsDir(givenFilename string) bool {
	fmt.Println("checkIsDir(" + givenFilename + ")\n")
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
