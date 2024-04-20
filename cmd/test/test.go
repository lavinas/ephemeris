package main

import (
	"fmt"
	"time"
)

func main() {
	dt := "10/02/2026"
	dtformat := "01/2006"
	local, _ := time.LoadLocation("America/Sao_Paulo")
	date, _ := time.ParseInLocation(dtformat, dt, local)
	fmt.Println(date)
}
