package main

import (
	"strings"
	"fmt"
)

func main() {
	x := "add client name ' paulo celso lavinas barbosa' document '044.179.328-24' responsible 'amelia cardoso' email 'lavinas@gmail.com' phone '11980876112' contact 'email'"
	y, _ := MapCommand(x)
	fmt.Print(y)
}

func MapCommand(cmd string) (map[string]string, error) {
	cmdSlice := strings.Split(strings.ToLower(cmd), " ")
	maps := make(map[string]string)
	lastKey := ""
	isParam := false
	param := ""

	for _, f := range cmdSlice {
		if f[0:1] == "'" || f[0:1] == "\"" {
			if len(f) > 1 {
				param := f[1:len(f)-1]
			}
		}

		if f[0:1] != "'" && f[0:1] != "\"" {
			lastKey = f
			maps[f] = ""
			continue
		}		
	}
	return maps, nil
}