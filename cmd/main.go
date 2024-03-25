package main

import (
	"fmt"

	"github.com/lavinas/ephemeris/pkg"
)

func main() {
	commands := pkg.NewCommands()
	var s = struct {
		Name string `command:"name: #name; not null"`
		Age  string `command:"name: #age"`
	}{}

	if err := commands.Unmarshal("#name test1 #age teste2", &s); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(s)
}
