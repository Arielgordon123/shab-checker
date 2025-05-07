package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"shab-checker/config"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type CellRange struct {
	StartRow, EndRow int
	StartCol, EndCol int
}

func NewSheetsClient(ctx context.Context, credentials config.Credentials) (*sheets.Service, error) {
	credsBytes, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}

	c, err := google.JWTConfigFromJSON(credsBytes, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, err
	}
	client := c.Client(ctx)
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return sheetsService, nil
}

func ReadSpreadsheet(service *sheets.Service, spreadsheetID string, sheetName string) ([][]interface{}, error) {
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, sheetName).Do()
	if err != nil {
		return nil, err
	}

	return resp.Values, nil
}

func GetSheetNames(ctx context.Context, service *sheets.Service, spreadsheetID string) ([]string, error) {
	resp, err := service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, err
	}

	sheetNames := make([]string, len(resp.Sheets))
	for i, sheet := range resp.Sheets {
		sheetNames[i] = sheet.Properties.Title
	}

	return sheetNames, nil
}

func AddSheet(service *sheets.Service, spreadsheetID, sheetTitle string, sheetRTL bool) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	request := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title:       sheetTitle,
						RightToLeft: sheetRTL,
					},
				},
			},
		},
	}

	return service.Spreadsheets.BatchUpdate(spreadsheetID, request).Do()
}

func WriteToSheet(service *sheets.Service, spreadsheetID, sheetName string, data [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: data,
	}

	_, err := service.Spreadsheets.Values.Update(spreadsheetID, sheetName, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}

	return nil
}

func ClearSheet(service *sheets.Service, spreadsheetID, sheetName string) error {
	clearRequest := &sheets.BatchClearValuesRequest{
		Ranges: []string{sheetName},
	}

	_, err := service.Spreadsheets.Values.BatchClear(spreadsheetID, clearRequest).Do()
	if err != nil {
		return err
	}

	return nil
}

func parseCellRef(sheetName, cellRef string) (col int, row int, err error) {

	for i, char := range cellRef {
		if unicode.IsDigit(char) {
			colPart := strings.ToUpper(cellRef[:i])

			rowStr := cellRef[i:]
			row, err := strconv.Atoi(rowStr)

			if err != nil {
				log.Println(err)
				return 0, 0, fmt.Errorf("invalid row number: %v", err)
			}

			col, err := parseCol(colPart)
			if err != nil {
				log.Println(err)
				return 0, 0, fmt.Errorf("invalid column reference in cell %s on sheet %s: %v", cellRef, sheetName, err)
			}
			// row, err := parseRow(rowPart)
			// 	if err != nil {
			// 		return 0, 0, fmt.Errorf("invalid row reference in cell %s on sheet %s: %v", cellRef, sheetName, err)
			// 	}
			return row - 1, col, nil
		}
	}
	return row - 1, col, nil

	// return 0, 0, fmt.Errorf("invalid input format: %s", sheetName)
}

// func parseCellRef(sheetName, cellRef string) (int, int, error) {
// 	if len(cellRef) < 2 {
// 		return 0, 0, fmt.Errorf("invalid cell reference: %s", cellRef)
// 	}

// 	colPart := strings.ToUpper(cellRef[:len(cellRef)-1])
// 	rowPart := cellRef[len(cellRef)-1:]
// 	log.Println("colpart", colPart)
// 	log.Println("rowPart", rowPart)
// 	col, err := parseCol(colPart)
// 	if err != nil {
// 		return 0, 0, fmt.Errorf("invalid column reference in cell %s on sheet %s: %v", cellRef, sheetName, err)
// 	}

// 	row, err := parseRow(rowPart)
// 	if err != nil {
// 		return 0, 0, fmt.Errorf("invalid row reference in cell %s on sheet %s: %v", cellRef, sheetName, err)
// 	}

// 	// Adjust row and column indices to zero-based
// 	return row - 1, col, nil
// }

func parseCol(colPart string) (int, error) {
	col := 0
	for _, char := range colPart {
		col = col*26 + int(char-'A')
	}
	return col, nil
}

func parseRow(rowPart string) (int, error) {
	row, err := strconv.Atoi(rowPart)
	if err != nil {
		return 0, err
	}
	return row, nil
}

func ParseCellRanges(sheetName string, cellRange string, timeRange string, titleRange string) ([]CellRange, error) {
	var parsedRanges []CellRange
	var ranges = []string{
		cellRange,
		timeRange,
		titleRange,
	}
	for _, c := range ranges {

		parts := strings.Split(c, ":")

		switch len(parts) {
		case 1:
			// Single cell reference
			row, col, err := parseCellRef(sheetName, parts[0])
			if err != nil {
				return nil, err
			}
			parsedRanges = append(parsedRanges, CellRange{
				StartRow: row,
				EndRow:   row,
				StartCol: col,
				EndCol:   col,
			})
		case 2:
			// Cell range
			startRow, startCol, err := parseCellRef(sheetName, parts[0])

			if err != nil {
				return nil, err
			}

			endRow, endCol, err := parseCellRef(sheetName, parts[1])

			if err != nil {
				return nil, err
			}

			parsedRanges = append(parsedRanges, CellRange{
				StartRow: startRow,
				EndRow:   endRow,
				StartCol: startCol,
				EndCol:   endCol,
			})
		default:
			return nil, fmt.Errorf("invalid cell range format: %s", c)

		}

	}
	return parsedRanges, nil
}
