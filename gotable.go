package gotable

/*
 * GoTable
 * Copyright (c) 2020-2021 Aaron Turner  <synfinatic at gmail dot com>
 *
 * This program is free software: you can redistribute it
 * and/or modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or with the authors permission any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strings"
)

const (
	TABLE_HEADER_TAG = "header"
	NOT_SUPPORTED    = "NO_SUPPORT"
)

type TableStruct interface {
	GetHeader(string) (string, error)
}

// Returns a row and a mapping of struct field name to header names
func TableRow(table TableStruct) (map[string]string, map[string]string, error) {
	row := map[string]string{}
	tbl := reflect.ValueOf(table)
	fieldCnt := tbl.Type().NumField()
	headers := make(map[string]string, fieldCnt)

	for i := 0; i < fieldCnt; i++ {
		f := reflect.TypeOf(table).Field(i)
		header, err := table.GetHeader(f.Name)
		if err != nil {
			return row, row, err
		}
		fval := tbl.FieldByName(f.Name)
		headers[f.Name] = header
		if !fval.IsValid() {
			continue // this shouldn't happen, but isn't fatal so ignore
		}
		switch fval.Kind() {
		case reflect.String:
			row[f.Name] = fval.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			row[f.Name] = fmt.Sprintf("%d", fval.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			row[f.Name] = fmt.Sprintf("%d", fval.Uint())
		case reflect.Bool:
			if fval.Bool() {
				row[f.Name] = "true"
			} else {
				row[f.Name] = "false"
			}
		default:
			// unsupported type!  so we mark it unsupported
			row[f.Name] = NOT_SUPPORTED
		}
	}
	return row, headers, nil
}

// Geneates a table using a list of TableStruct & struct field names in the report
func GenerateTable(tables []TableStruct, fields []string) error {
	table := []map[string]string{}
	headers := map[string]string{}
	for _, item := range tables {
		row, h, err := TableRow(item)
		if err != nil {
			return err
		}
		table = append(table, row)
		headers = h
	}

	generateTable(table, headers, fields)
	return nil
}

// Generates a CSV output instead of a table- no header
func GenerateCSV(tables []TableStruct, fields []string) error {
	table := []map[string]string{}
	for _, item := range tables {
		row, _, err := TableRow(item)
		if err != nil {
			return err
		}
		table = append(table, row)
	}

	generateCSV(table, fields)
	return nil
}

func generateTable(data []map[string]string, fieldMap map[string]string, fields []string) {
	table := [][]string{}
	colWidth := make([]int, len(fields))

	// figure out width of column headers
	for i, field := range fields {
		colWidth[i] = len(fieldMap[field])
	}

	// calc max len of every column & build our row
	for _, r := range data {
		row := make([]string, len(fields))
		for i, field := range fields {
			row[i] = r[field]
			if len(r[field]) > colWidth[i] {
				colWidth[i] = len(r[field])
			}
		}
		table = append(table, row)
	}

	// build our fstring for each row
	fstrings := make([]string, len(fields))
	for i, width := range colWidth {
		fstrings[i] = fmt.Sprintf("%%-%ds", width)
	}
	fstring := strings.Join(fstrings, " | ")
	fstring = fmt.Sprintf("%s\n", fstring)

	// fmt.Sprintf() expects []interface...
	finter := make([]interface{}, len(fields))
	for i, field := range fields {
		finter[i] = fieldMap[field]
	}

	// print the header
	headerLine := fmt.Sprintf(fstring, finter...)
	fmt.Printf("%s%s\n", headerLine, strings.Repeat("=", len(headerLine)-1))

	// print each row
	for _, row := range data {
		values := make([]interface{}, len(fields))
		for i, field := range fields {
			values[i] = row[field]
		}
		fmt.Printf(fstring, values...)
	}
}

func generateCSV(data []map[string]string, fields []string) error {
	var err error
	fStr := make([]string, len(fields))
	for i, _ := range fields {
		fStr[i] = "%s"
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	for _, row := range data {
		values := make([]string, len(fields))
		for i, field := range fields {
			values[i] = row[field]
		}
		if err = w.Write(values); err != nil {
			return err
		}
	}
	return err
}

func GetHeaderTag(v reflect.Value, fieldName string) (string, error) {
	field, ok := v.Type().FieldByName(fieldName)
	if !ok {
		return "", fmt.Errorf("Invalid field '%s' in %s", fieldName, v.Type().Name())
	}
	tag := string(field.Tag.Get(TABLE_HEADER_TAG))
	return tag, nil
}
