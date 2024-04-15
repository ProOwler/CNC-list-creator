package main

import (
	"fmt"
	//"io"
	"log"
	"os"
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
	// fmt.Println(i2)
	var homePath string
	homePath, dirErr := os.UserHomeDir()
	if dirErr == nil {
		fmt.Println(homePath)
		var myText string = "Домашняя папка: \n" + homePath
		err := os.WriteFile("hello.txt", []byte(myText), 0666)
		if err != nil {
			log.Fatal(err)
		}
		if err == nil {
			fmt.Println("Успешно!")
		}
	}
}
