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
	// nameForListOfFiles := "list.xml"
	nameForListOfFiles := "output.txt"

	// чтоб компилятор не ругался на неиспользование массива
	listOfFileFormats := []string{"xml", "XML", "mpr", "MPR"}
	if len(listOfFileFormats) > 1 {
	}

	check(os.WriteFile(nameForListOfFiles, []byte(getOutput()), 0666))
	fmt.Println("Успешно!")
}

/*
func recursiveWalkthrough(startPath string, outputFilename string, fileFormats []string, ) {
	s
}
*/

func getOutput() string {
	absoluteFilepath, err := filepath.Abs(getMainStartupArg())
	check(err)
	return outputToOneString(getListOfDirAndFiles(absoluteFilepath))
}

func getMainStartupArg() string {
	if len(os.Args) > 1 {
		fmt.Println("Скормлена папка: " + os.Args[1])
		return (os.Args[1])
	} else {
		fmt.Println("Вызван сам по себе: " + os.Args[0])
		return os.Args[0]
	}
}

func outputToOneString(strings []string) string {
	myString := ""
	myOutput := strings
	for i := 0; i < len(myOutput); i++ {
		myString += myOutput[i] + "\n"
	}
	return myString
}

func getListOfDirAndFiles(givenFilename string) []string {
	myFile, err1 := os.Open(getDirOfArg(givenFilename))
	check(err1)

	myList, err1 := myFile.Readdirnames(-1)
	check(err1)

	return myList
}

func getDirOfArg(givenFilename string) string {
	myFileInfo, err1 := os.Stat(givenFilename)
	check(err1)

	mydirPath := givenFilename
	if !myFileInfo.IsDir() {
		mydirPath = filepath.Dir(givenFilename)
	}
	return mydirPath
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}
