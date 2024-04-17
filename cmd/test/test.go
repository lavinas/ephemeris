package main

import (
	"fmt"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

func main() {
	dto := dto.ClientAddIn{}
	input := "client add date 01/06/2020 id paulo name Paulo Lavinas email lavinas@gmail.com phone 11999999999 document 12345678901 role client ref ref1"

	commands := pkg.NewCommands2()
	if err := commands.Unmarshal(input, &dto); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(1, dto.Action)
	fmt.Println(2, dto.Object)
	fmt.Println(3, dto.ID)
	fmt.Println(4, dto.Date)
	fmt.Println(5, dto.Name)
	fmt.Println(6, dto.Email)
	fmt.Println(7, dto.Phone)
	fmt.Println(8, dto.Document)
	fmt.Println(9, dto.Role)
	fmt.Println(10, dto.Ref)

}
