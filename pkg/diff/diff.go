package diff

import (
	"fmt"
	"log"
	"shab-checker/config"
	"shab-checker/pkg/sheets"
	"strings"
)

type Cell struct {
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Value string `json:"value"`
	Time  string `json:"time"`
	Title string `json:"title"`
}

type ChangedCell struct {
	Date     string `json:"date"`
	Cell     `json:"cell"`
	OldValue string `json:"oldValue"`
}

func CompareSpreadsheetsAndGetDiff(sheet1, sheet2 [][]interface{}, sheetName string, preDefinedCells []config.PreDefinedCells) []ChangedCell {
	var changedCells []ChangedCell

	// Create a map for sheet1 to efficiently lookup cells
	sheet1Map := make(map[Cell]string)
	for _, cellConfig := range preDefinedCells {

		cellRange, err := sheets.ParseCellRanges(sheetName, cellConfig.CellRange, cellConfig.TitleRange, cellConfig.TimeRange)

		if err != nil {
			log.Printf("Error parsing cell ranges for sheet %s: %v", sheetName, err)
			continue
		}

		for row := cellRange[0].StartRow; row <= cellRange[0].EndRow; row++ {
			for col := cellRange[0].StartCol; col <= cellRange[0].EndCol; col++ {
				if row < len(sheet1) && col < len(sheet1[row]) {

					// if row < len(sheet1) && col < len(sheet1[row]) {

					value := interfaceToString(sheet1[row][col])
					title := ""
					time := ""
					if cellRange[2].StartRow < len(sheet1) && cellRange[2].StartCol < len(sheet1[cellRange[2].StartRow]) {

						time = interfaceToString(sheet1[cellRange[2].StartRow][cellRange[2].StartCol])
					}
					if cellRange[1].StartRow < len(sheet1) && cellRange[1].StartCol < len(sheet1[cellRange[1].StartRow]) {

						title = interfaceToString(sheet1[cellRange[1].StartRow][cellRange[1].StartCol])
					}
					cell := Cell{row, col, value, time, title}

					sheet1Map[cell] = cell.Value

					// Check if the corresponding cell exists in sheet2
					if row < len(sheet2) && col < len(sheet2[row]) {
						if interfaceToString(sheet2[row][col]) != cell.Value {
							changedCells = append(changedCells, ChangedCell{sheetName, cell, interfaceToString(sheet2[row][col])})
						}
					} else {
						// If the cell doesn't exist in sheet2, consider it as a new cell

						changedCells = append(changedCells, ChangedCell{sheetName, cell, ""})
					}
				}
			}
		}
	}

	return changedCells
}

func interfaceToString(value interface{}) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}
