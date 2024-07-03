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
	if len(x) != len(y) {
		return 0, fmt.Errorf("x, y must have the same length")
	}
	if w != nil && len(w) < len(x) {
		return 0, fmt.Errorf("w must be at least the same length as x")
	}
	k := 1.00000000
	dec := 1 / float64(len(x))
	z := 0.00000000
	for i := 0; i < len(x); i++ {
		z1 := float64(x[i] - y[i])
		z1 = z1 * z1
		if w != nil {
			z1 = z1 * w[i]
		} else {
			z1 = z1 * k
			k -= dec
		}
		z += z1
	}
	return math.Sqrt(z), nil
}