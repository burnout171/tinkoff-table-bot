package main

import (
	"fmt"
	"strings"
	"strconv"
	"time"

	"google.golang.org/api/sheets/v4"
)

// TableManagement manages update and get table data commands
type TableManagement struct {
	ts *TableService
}

// NewTableManagement creates new TableManagement instant
func NewTableManagement(ts *TableService) *TableManagement {
	tm := &TableManagement{}
	tm.ts = ts
	return tm
}

// GetTableBalance returns kind of balance for current month
func (tm *TableManagement) GetTableBalance(command string) (string, error) {
	switch command {
	case "db":
		return tm.getDailyBalance()
	case "mb":
		return tm.getMonthlyBalance()
	case "ma":
		return tm.getMonthlyAccumulation()
	default:
		return "Unknown command", nil
	}
}

// UpdateTableData updates daily balance and returns its new value
func (tm *TableManagement) UpdateTableData(input string) (int64, error) {
	receivedKey, sum := tm.parseInput(input)
	month, day := tm.currentDate()
	workingRange := fmt.Sprintf("%s!H%d:I%d", month, day+1, day+1)
	receivedRange, err := tm.ts.GetData(workingRange)
	if err != nil {
		return -1, err
	}
	var resultRange sheets.ValueRange
	var myValues []interface{}
	if len(receivedRange.Values) == 0 {
		myValues = []interface{}{strings.ToLower(receivedKey), sum}
	} else {
		key := tm.prepareKey(receivedKey, receivedRange.Values[0][0].(string))
		value := tm.prepareValue(sum, receivedRange.Values[0][1].(string))
		myValues = []interface{}{strings.ToLower(key), value}
	}
	resultRange.Values = append(resultRange.Values, myValues)
	updateResponse, err := tm.ts.UpdateData(workingRange, &resultRange)
	if (err != nil) {
		return -1, err
	}
	return updateResponse.UpdatedCells, nil
}

func (tm *TableManagement) getDailyBalance() (string, error) {
	month, day := tm.currentDate()
	workingRange := fmt.Sprintf("%s!K%d", month, day+1)
	return tm.getSimpleSheetData(workingRange)
}

func (tm *TableManagement) getMonthlyBalance() (string, error) {
	month, _ := tm.currentDate()
	workingRange := fmt.Sprintf("%s!K33", month)
	return tm.getSimpleSheetData(workingRange)
}

func (tm *TableManagement) getMonthlyAccumulation() (string, error) {
	month, _ := tm.currentDate()
	workingRange := fmt.Sprintf("%s!D21", month)
	return tm.getSimpleSheetData(workingRange)
}

func (tm *TableManagement) getSimpleSheetData(workingRange string) (string, error) {
	receivedRange, err := tm.ts.GetData(workingRange)
	if err != nil {
		return "", err
	}
	return receivedRange.Values[0][0].(string), nil
}

func (tm *TableManagement) currentDate() (monthName string, day int) {
	_, month, day := time.Now().Date()
	switch month {
	case time.January:
		return "Январь", day
	case time.February:
		return "Февраль", day
	case time.March:
		return "Март", day
	case time.April:
		return "Апрель", day
	case time.May:
		return "Май", day
	case time.June:
		return "Июнь", day
	case time.July:
		return "Июль", day
	case time.August:
		return "Август", day
	case time.September:
		return "Сентябрь", day
	case time.October:
		return "Октябрь", day
	case time.November:
		return "Ноябрь", day
	case time.December:
		return "Декабрь", day
	default:
		return "", day
	}
}

func (tm *TableManagement) parseInput(input string) (description string, sum float64) {
	splitted := strings.Split(input, " ")
	var descriptionSlice []string
	for _, word := range splitted {
		if value, err := strconv.ParseFloat(word, 64); err == nil {
			sum += value
			continue
		}
		descriptionSlice = append(descriptionSlice, word)
	}
	return strings.Join(descriptionSlice, ", "), sum
}

func (tm *TableManagement) prepareKey(receivedKey string, currentKey string) string {
	if strings.Contains(currentKey, " + ") {
		currentKey = strings.Replace(currentKey, " + ", ", ", -1)
	}
	if receivedKey == "" {
		return currentKey
	}
	return currentKey + ", " + receivedKey
}

func (tm *TableManagement) prepareValue(sum float64, currentValue string) float64 {
	if strings.HasPrefix(currentValue, "SUM") { // In case SUM() function is used in Sheet to sum the exchanges
		currentValue = currentValue[4 : len(currentValue)-2]
		for _, number := range strings.Split(currentValue, ", ") {
			if value, err := strconv.ParseFloat(number, 64); err == nil {
				sum += value
				continue
			}
		}
		return sum
	}
	currentValue = strings.ReplaceAll(currentValue, ",", "")
	floatValue, _ := strconv.ParseFloat(currentValue, 64)
	return sum + floatValue
}