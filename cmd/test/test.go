package main

import (
	"fmt"

	td "github.com/masatana/go-textdistance"
)

var (
	x = "leandro-canto-2024-01-01-15-00"
	y = "leandro-piano-2024-01-01-00-00"
	z = "leandro-canto-2024-12-01-00-00"
)
// main is the entry point of the application
func main() {
	fmt.Println(td.DamerauLevenshteinDistance(x, y))
	fmt.Println(td.DamerauLevenshteinDistance(x, z))
	fmt.Println(td.JaroDistance(x, y))
	fmt.Println(td.JaroDistance(x, z))
}