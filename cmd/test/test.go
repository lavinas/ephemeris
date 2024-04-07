package main

import (
	"fmt"
	"strings"
)

func trim(table [][]string) [][]string {
	fillCols := map[int]bool{}
	fillLines := map[int]bool{}
	for i := 1; i < len(table); i++ {
		for j := 0; j < len(table[i]); j++ {
			if table[i][j] != "" {
				fillCols[j] = true
				fillLines[i] = true
			}
		}
	}
	ret := make([][]string, 0)
	line := []string{}
	for i := 0; i < len(table[0]); i++ {
		if fillCols[i] {
			line = append(line, table[0][i])
		}
	}
	if len(line) > 0 {
		ret = append(ret, line)
	}
	for i := 1; i < len(table); i++ {
		if !fillLines[i] {
			continue
		}
		line := []string{}
		for j := 0; j < len(table[i]); j++ {
			if fillCols[j] {
				line = append(line, table[i][j])
			}
		}
		if len(line) > 0 {
			ret = append(ret, line)
		}
	}
	return ret
}

func MountTable(table [][]string) string {
	if len(table) == 0 || len(table[0]) == 0{
		return ""
	}
	ret := ""
	// get number of columns from the first table row
	columnLengths := make([]int, len(table[0]))
	for _, line := range table {
		for i, val := range line {
			if len(val) > columnLengths[i] {
				columnLengths[i] = len(val)
			}
		}
	}
	var lineLength int
	for _, c := range columnLengths {
		lineLength += c + 3
	}
	lineLength += 1
	for i, line := range table {
		if i == 0 {
			ret += fmt.Sprintf("+%s+\n", strings.Repeat("-", lineLength-2))
		}
		for j, val := range line {
			ret += fmt.Sprintf("| %-*s ", columnLengths[j], val)
			if j == len(line)-1 {
				ret += "|\n"
			}
		}
		if i == 0 || i == len(table)-1 {
			ret += fmt.Sprintf("+%s+\n", strings.Repeat("-", lineLength-2))
		}
	}
	return ret
}



func main () {
	table := [][]string{
		{"a", "b", "c"},
		{"1", "1", "1"},
		{"2", "", ""},
		{"", "", ""},
		{"", "", ""},
		{"", "", ""},
	}
	// table = DelBlankCols(table)
	fmt.Println(MountTable(table))
	ret := trim(table)
	fmt.Println(MountTable(ret))
}