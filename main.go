package main

import (
	"fmt"
	//"io"
	"log"
	"os"
	//"path"
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
	check(os.WriteFile("output.txt", []byte(outputToOneString(getOutput())), 0666))
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
	absoluteFilepathToTheExecutable, err := filepath.Abs(os.Args[0])
	check(err)
	return getDirAndFiles(absoluteFilepathToTheExecutable)

}

func getDirAndFiles(s string) []string {
	myFile, err1 := os.Open(s)
	check(err1)

	myFilePath := filepath.Dir(myFile.Name())
	myFile, err1 = os.Open(myFilePath)
	check(err1)

	myList, err1 := myFile.Readdirnames(-1)
	check(err1)

	return append([]string{myFilePath}, myList...)
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}
