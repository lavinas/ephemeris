package main

import (
	"fmt"
	"math"

	// td "github.com/masatana/go-textdistance"
)

var (
	x = "2024-01-01-15-00-canto"
	y = "2024-01-01-00-00-canto"
	z = "2024-01-01-15-00-piano"
)
// main is the entry point of the application
func main() {
	fmt.Println(WeigthedEuclidean(x, y, nil))
	fmt.Println(WeigthedEuclidean(x, z, nil))
}


// WeigthedDistance is a custom string euclidean distance between x, y
// and w is a slice of weights for each character in x and y
// if w is nil, the weights are calculated as a decreasing sequence
// starting in 1 and ending in 1/len(x)
// returns the euclidean distance between x and y with weights w or an error
func WeigthedEuclidean(x, y string, w []float64) (float64, error) {
	if w == nil {
		w = []float64{}
		k := 1.00000000
		z := 1 / float64(len(x))
		for i := 0; i < len(x); i++ {
			w = append(w, k)
			k -= z		
		}
	}
	if len(x) != len(y) || len(x) != len(w) {
		return 0, fmt.Errorf("x, y and w must have the same length")
	}
	z := 0.00000000
	for i := 0; i < min(len(x), len(y), len(w)); i++ {
		z += float64(x[i] - y[i]) * float64(x[i] - y[i]) * w[i]
	}
	return math.Sqrt(z), nil
}