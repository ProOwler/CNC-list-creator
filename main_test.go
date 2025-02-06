package main

import (
	"testing"
)

func TestHasStringInList(t *testing.T) {
	// Arrange - подготовка всех входных данных и ожидаемого результата
	const requiredStrGO, requiredStrGDO = "Go", "Gdo"
	var stringList = []string{"peach", "Gopher", "apple", "Gdo"}
	//sort.Strings(stringList)
	const wantF = false
	const wantT = true
	// Action - вызов тестируемой функции с эталонными параметрами
	gotGO := hasStringInList(requiredStrGO, stringList)
	gotGDO := hasStringInList(requiredStrGDO, stringList)
	// Assert - сравнение полученного значения с ожидаемым и вывод сообщения, если они не совпадают
	if gotGO != wantF {
		t.Errorf("hasStringInList(%q, %q); \ngot %t; \nwant %t", requiredStrGO, stringList, gotGO, wantF)
	}
	if gotGDO != wantT {
		t.Errorf("hasStringInList(%q, %q); \ngot %t; \nwant %t", requiredStrGDO, stringList, gotGDO, wantT)
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
