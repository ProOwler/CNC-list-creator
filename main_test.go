package main

import (
	"testing"
)

func TestHasStringInList(t *testing.T) {
	// Arrange - подготовка всех входных данных и ожидаемого результата
	const requiredStr = "Go"
	var stringList = []string{"peach", "Gopher", "apple", "Gdo", "pear", "plum"}
	//sort.Strings(stringList)
	const want = false
	// Action - вызов тестируемой функции с эталонными параметрами
	got := hasStringInList(requiredStr, stringList)
	// Assert - сравнение полученного значения с ожидаемым и вывод сообщения, если они не совпадают
	if got != want {
		t.Errorf("hasStringInList(%q, %q) = %t; \nwant %t", requiredStr, stringList, got, want)
	}
}

func TestGetExtention(t *testing.T) {
	// Arrange
	var testStrs = []string{"Go", "Gopher.go", "C:\\apple", "C:\\plum.txt"}
	var wantStrs = []string{"", "go", "", "txt"}
	// Action
	for i := 0; i < len(testStrs); i++ {
		got := getExtention(testStrs[i])
		want := wantStrs[i]
		// Assert
		if got != want {
			t.Errorf("got = %v; \nwant = %v", got, want)
		}
	}
}
