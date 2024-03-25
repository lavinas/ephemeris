package main

import (
	"fmt"

	"github.com/lavinas/ephemeris/pkg"
)

func main() {
	commands := pkg.NewCommands()
	var s = struct {
		Name string `command:"name:#name; not null"`
		Age  string `command:"name:#age; key"`
	}{}

	if err := commands.Unmarshal("#name #age 2222", &s); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(s)
}
