package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

var (
	x = []interface{}{
		int64(20),
		"piano_10",
		"piano_20",
		"piano_30",
	}
	y = []interface{}{
		int64(10),
		"piano_10",
		"piano_20",
		"piano_30",
	}
)

// main is the entry point of the application
func main() {
	d, err := WeightedDistance(x, y, []float64{1, 1, 1, 1})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(d)
}

// Distance calculates the distance between two slices of interfaces
// interface{} can be any type of data, but implemented only for strings, time, int64 and float64
// w is the weight for each field
// returns the distance between the two sliceswhen 0 means that the two slices are equal
func WeightedDistance(x, y []interface{}, w []float64) (float64, error) {
	if len(x) != len(y) || len(x) != len(w) {
		return 0, fmt.Errorf("slices must have the same length")
	}
	z := 0.0
	total := 0.0
	for i := range x {
		if reflect.TypeOf(x[i]) != reflect.TypeOf(y[i]) {
			return 0, fmt.Errorf("slices must have the same types")
		}
		z1 := 0.0
		switch reflect.TypeOf(x[i]).String() {
		case "string":
			z1 = float64(levenshtein.DistanceForStrings([]rune(x[i].(string)), []rune(y[i].(string)), levenshtein.DefaultOptions))
		case "int64":
			z1 = float64(x[i].(int64) - y[i].(int64))
		case "float64":
			z1 = x[i].(float64) - y[i].(float64)
		case "time.Time":
			z1 = float64(x[i].(time.Time).Sub(y[i].(time.Time)).Seconds())
		default:
			fmt.Println(reflect.TypeOf(x[i]).String())
			return 0, fmt.Errorf("type not implemented. Implemented: string, int64, float64, time.Time")
		}
		z += z1 * w[i]
		total += w[i]
	}
	return z / total, nil
}
