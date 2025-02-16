package main

import (
	"fmt"
	"testing"
	"time"
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

func TestGetUpdatedXML(t *testing.T) {
	// Arrange
	tThen := time.Now()
	var testStrs = `<?xml version="1.0" encoding="utf-8" ?>
<Root>
  <Project Name="" Flag="SWJ008">
    <Panels>
      <Panel ID="1.4_4_Ковыряло_левое_(фрезеровка)" Name="740_432_1" Width="432.250" Length="740.500" Material="" Thickness="16.000" IsProduce="true" MachiningPoint="1" Type="1" Face5ID="" Face6ID="" Grain="L" Count="1">
        <Machines>
          <Machining ID="1000" Type="3" IsGenCode="2" Face="5" Depth="16.000" X="670.000" Y="25.000" Pocket="0" ToolOffset="左">
            <Lines>
              <Line LineID="1" EndX="715.000" EndY="25.000" Angle="0.000000" />
              <Line LineID="2" EndX="725.000" EndY="35.000" Angle="-90.000" />
            </Lines>
          </Machining>
        </Machines>
        <EdgeGroup>
          <Edge Face="1" Thickness="0.000000" />
          <Edge Face="2" Thickness="0.000000" />
        </EdgeGroup>
      </Panel>
    </Panels>
  </Project>
</Root>`
	var wantStrs = `<?xml version="1.0" encoding="utf-8" ?>
<Root>
  <Project Name="" Flag="SWJ008">
    <Panels>
      <Panel ID="1.4_4_Ковыряло_левое_(фрезеровка)" Name="740.5_432.2" Width="432.250" Length="740.500" Material="" Thickness="16.000" IsProduce="true" MachiningPoint="1" Type="1" Face5ID="" Face6ID="" Grain="L" Count="1">
        <Machines>
          <Machining ID="1000" Type="3" IsGenCode="2" Face="5" Depth="16.000" X="670.000" Y="25.000" Pocket="0" ToolOffset="左">
            <Lines>
              <Line LineID="1" EndX="715.000" EndY="25.000" Angle="0.000000" />
              <Line LineID="2" EndX="725.000" EndY="35.000" Angle="-90.000" />
            </Lines>
          </Machining>
        </Machines>
        <EdgeGroup>
          <Edge Face="1" Thickness="0.000000" />
          <Edge Face="2" Thickness="0.000000" />
        </EdgeGroup>
      </Panel>
    </Panels>
  </Project>
</Root>`
	// Action
	got, err := getUpdatedXML(testStrs)
	want := wantStrs
	// Assert
	if err != nil {
		t.Errorf("Ошибка: %v", err)
	}
	if got != want {
		t.Errorf("got = %v; \nwant = %v", got, want)

	}
	fmt.Printf("Elapsed %.6f sec", time.Since(tThen).Seconds())
	fmt.Println("")
}
