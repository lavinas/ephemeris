package main

import (
	"github.com/lavinas/ephemeris/pkg"
)

var (
	x = struct {
		Name string `command:"name:name; key; not null"`
		Name2 string `command:"name:name; key; not null"`
	}{}

)

func main() {
	com := pkg.Commands{}
	err := com.Unmarshal("name alex age 20 mood test other test2 another xxx", &x)
	if err != nil {
		println(err.Error())
	}
	println(com.Marshal(&x))
}
