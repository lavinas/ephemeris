package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lavinas/ephemeris/internal/adapters/handler"
	"github.com/lavinas/ephemeris/internal/adapters/repository"
	"github.com/lavinas/ephemeris/internal/usecase"
)

// main is the entry point of the application
func main() {
	repo, err := repository.NewRepository(os.Getenv("MYSQL_DNS"))
	if err != nil {
		fmt.Println("internal error: " + err.Error())
		return
	}
	logger := log.New(os.Stdout, "ephemeris: ", log.LstdFlags)
	usecase := usecase.NewUsecase(repo, logger)
	handler := handler.NewCommandHandler(usecase)
	handler.Run()
	fmt.Println("done")

}
