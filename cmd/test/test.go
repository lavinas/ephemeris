package main

import (
	"fmt"
	"strings"

)

func main() {
	x := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	fmt.Println(strings.Join(x, ", "))
}
