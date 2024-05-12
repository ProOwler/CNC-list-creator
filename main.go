package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func main() {
	// myString := outputToOneString(getOutput())
	// nameForListOfFiles := "list.xml"
	nameForListOfFiles := "output.txt"

	check(os.WriteFile(nameForListOfFiles, []byte(outputToOneString(getOutput())), 0666))
	fmt.Println("Успешно!")
}

func outputToOneString(strings []string) string {
	myString := ""
	myOutput := strings
	for i := 0; i < len(myOutput); i++ {
		myString += myOutput[i] + "\n"
	}
	return myString
}

func getOutput() []string {
	absoluteFilepath, err := filepath.Abs(getAbsoluteFilepathFromStartupArgs())
	check(err)
	return getListOfDirAndFiles(absoluteFilepath)
}

func getListOfDirAndFiles(givenFilename string) []string {
	myFileInfo, err1 := os.Stat(givenFilename)
	check(err1)

	mydirPath := givenFilename
	if !myFileInfo.IsDir() {
		mydirPath = filepath.Dir(givenFilename)
	}

	myFile, err1 := os.Open(mydirPath)
	check(err1)

	myList, err1 := myFile.Readdirnames(-1)
	check(err1)

	return myList
}

func getAbsoluteFilepathFromStartupArgs() string {
	//return "D:\\Prowler\\projects\\CNC-list-creator\\proj\\.git\\"
	if len(os.Args) > 1 {
		fmt.Println("Скормлена папка: " + os.Args[1])
		return (os.Args[1])
	} else {
		fmt.Println("Вызван сам по себе: " + os.Args[0])
		return os.Args[0]
	}
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}
