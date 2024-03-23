package main

import (
	"fmt"

	"github.com/lavinas/ephemeris/pkg"
)

func main() {
	commands := pkg.NewCommands()
	var s = struct {
		Name string `json:"name"`
		Age  string    `json:"age"`
	}{}

	if err := commands.Unmarshal("", &s, "json"); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(s)
}
