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
	// client := domain.Client{Base: domain.Base{ID: "paulo"}}
	client := domain.Client{ID: "paulo"}
	clients := []domain.Client{}

	err = repo.Find(&client, &clients)
	if err != nil {
		panic(err)
	}
	for _, c := range clients {
		fmt.Println(c)
	}
}
