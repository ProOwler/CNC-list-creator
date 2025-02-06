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
