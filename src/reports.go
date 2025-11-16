package main

import (
	"encoding/xml"
	"log"
	"os"
	"sort"
	"strings"
)

// XML-представление отчёта
type XReportHead struct {
	XMLName        xml.Name        `xml:"Root"`
	ReportItemList XReportItemList `xml:"ReportItemList"`
}

type XReportItemList struct {
	ReportItem []XReportItem `xml:"ReportItem"`
}

type XReportItem struct {
	ItemName       string          `xml:"ItemName,attr"`
	Status         string          `xml:"Status,attr"`
	DateReady      string          `xml:"DateReady,attr"`
	Level          int             `xml:"Level,attr"`
	ReportItemList XReportItemList `xml:"ReportItemList,omitempty"`
}

// GO-представление отчёта
type ReportObj struct {
	itemName   string
	status     string
	dateReady  string
	level      int
	innerItems []ReportObj
}

func createReport(reports []ReportObj) string {
	var reportStrings []string
	for _, rep := range reports {
		dateMonth := ""
		if rep.dateReady != "" {
			dateMonth = rep.dateReady[0:7]
		}
		reportStrings = append(reportStrings, dateMonth+" - "+rep.itemName+"\n")
	}
	sort.Strings(reportStrings)
	var sb strings.Builder
	for _, rs := range reportStrings {
		sb.WriteString(rs)
	}
	return sb.String()
}

func getReportObjectsFromFile(fullFileName string) []ReportObj {
	myFileBytes, err := os.ReadFile(fullFileName)
	if err != nil {
		log.Printf("Не удалось прочитать файл отчёта %s: %w\n", fullFileName, err)
		return []ReportObj{{}}
	}
	var myRepXML XReportHead
	err = xml.Unmarshal(myFileBytes, &myRepXML)
	if err != nil {
		log.Printf("Не удалось разобрать XML из файла отчёта %s: %w\n", fullFileName, err)
		return []ReportObj{{}}
	}
	return getReportObjects(myRepXML)
}

func (item *ReportObj) writeReportToFile(fullFilePath string) {
	var objects []ReportObj
	objects = append(objects, *item)
	xmlReport := getReportXML(objects)
	myHeader := `<?xml version="1.0" encoding="utf-8" ?>` + "\n"
	xmlReportString := ""
	xmlReportBytes, errMarshal := xml.MarshalIndent(xmlReport, "", "	") // Используем табуляцию для отступов
	if errMarshal != nil {
		log.Printf("Ошибка при сериализации XML: %v", errMarshal)
		return
	}

	xmlReportString = myHeader + string(xmlReportBytes)
	createFile(fullFilePath, []byte(xmlReportString))
}

func getReportXML(itemObj []ReportObj) XReportHead {
	var result = XReportHead{}
	for _, entry := range itemObj {
		result.ReportItemList.ReportItem = append(result.ReportItemList.ReportItem, entry.convertReportItemToXML())
	}

	return result
}

func (item *ReportObj) convertReportItemToXML() XReportItem {
	var result = XReportItem{
		ItemName:  item.itemName,
		Level:     item.level,
		DateReady: item.dateReady,
		Status:    item.status,
	}
	for _, entry := range item.innerItems {
		result.ReportItemList.ReportItem = append(result.ReportItemList.ReportItem, entry.convertReportItemToXML())
	}
	return result
}

func getReportObjects(itemX XReportHead) []ReportObj {
	var result = []ReportObj{}
	for _, entry := range itemX.ReportItemList.ReportItem {
		result = append(result, entry.convertReportItemToObj())
	}
	return result
}

func (item *XReportItem) convertReportItemToObj() ReportObj {
	var result = ReportObj{
		itemName:  item.ItemName,
		level:     item.Level,
		dateReady: item.DateReady,
		status:    item.Status,
	}
	for _, entry := range item.ReportItemList.ReportItem {
		result.innerItems = append(result.innerItems, entry.convertReportItemToObj())
	}
	return result
}
