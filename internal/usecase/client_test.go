package usecase

import (
	"log"
	"testing"
	"os"

	"github.com/lavinas/ephemeris/internal/adapter/repository"
	"github.com/lavinas/ephemeris/internal/domain"
)

func TestClientOk(t *testing.T) {
	dns := os.Getenv("MYSQL_DNS")
	repo, error := repository.NewRepository(dns)
	if error != nil {
		t.Errorf("TestAddClient failed: %s", error)
	}
	if err := repo.Migrate([]interface{}{&domain.Client{}}); err != nil {
		t.Errorf("TestAddClient failed: %s", err)
	}
	log := log.New(os.Stdout, "test: ", log.LstdFlags)
	usecase := NewClientUsecase(repo, log)
	// Test ok
	repo.Delete(&domain.Client{}, "1")
	err := usecase.AddClient("1", "Test test", "Test test", "test@test.com", "11980876112", "email", "04417932824")
	if err != nil {
		t.Errorf("TestAddClient failed: %s", err)
	}
	cli, err := usecase.GetClient("1")
	if err != nil {
		t.Errorf("TestAddClient failed: %s", err)
	}
	if cli.Name != "Test test" {
		t.Errorf("TestAddClient failed: %s", cli.Name)
	}
	repo.Delete(&domain.Client{}, "1")
}
