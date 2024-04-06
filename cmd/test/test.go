package main

import (
	"fmt"
	"os"

	"github.com/lavinas/ephemeris/internal/adapters/repository"
	"github.com/lavinas/ephemeris/internal/domain"
)

func Insert() {
	repo, err := repository.NewRepository(os.Getenv("MYSQL_DNS"))
	if err != nil {
		fmt.Println("internal error: " + err.Error())
		return
	}
	defer repo.Close()
	client := domain.NewClient("paulo5", "01/01/2024", "Paulo Lavinas", "lavinas@gmail.com", "11908080808", "04417932824", "email")
	clientRole := domain.NewClientRole("paulo5", "01/01/2024", "paulo3", "client", "paulo3")
	if err = client.Format(); err != nil {
		fmt.Println("2 internal error: " + err.Error())
		return
	}
	if err = repo.Begin(); err != nil {
		fmt.Println("3 internal error: " + err.Error())
		return
	}
	if err = repo.Add(&client); err != nil {
		fmt.Println("4 internal error: " + err.Error())
		return
	}
	if err = repo.Add(&clientRole); err != nil {
		fmt.Println("4 internal error: " + err.Error())
		return
	}
	if err = repo.Commit(); err != nil {
		fmt.Println("5 internal error: " + err.Error())
		return
	}
	fmt.Println("6 client added")
}

func main() {
	Insert()
}
