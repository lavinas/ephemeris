package main

import (
	"fmt"
	"os"
)

func main() {
	x := os.Getenv("MYSQL_DNS")
	fmt.Println(x)
}