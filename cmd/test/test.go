package main

import (
	"fmt"
	"os"

	"github.com/lavinas/ephemeris/internal/adapters/repository"
	"github.com/lavinas/ephemeris/internal/domain"
)


func main() {
	repo, err := repository.NewRepository(os.Getenv("MYSQL_DNS"))
	if err != nil {
		panic(err)
	}
	defer repo.Close()
	client := domain.Client{Base: domain.Base{ID: "paulo"}}
	x, err := repo.Search(&client)
	if err != nil {
		panic(err)
	}
	fmt.Println(x)
}
