package main

import (
	"fmt"
	"strings"

	"github.com/lavinas/ephemeris/pkg"
)

var (
	cycles = []string{
		pkg.RecurrenceCycleOnce,
		pkg.RecurrenceCycleDay,
		pkg.RecurrenceCycleWeek,
		pkg.RecurrenceCycleMonth,
		pkg.RecurrenceCycleYear,
	}
)

func main() {
	x := strings.Join(cycles, ", ")
	fmt.Println(x)
}
