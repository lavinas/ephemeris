package usecase

import (
	"log"
	"testing"
	"os"

	"github.com/lavinas/ephemeris/internal/adapter/repository"
	"github.com/lavinas/ephemeris/internal/domain"
)

func TestClient(t *testing.T) {
	// prepare
	usecase := getUsecase(t)
	// add a client ok
	usecase.Repo.Delete(&domain.Client{}, "1")
	defer usecase.Repo.Delete(&domain.Client{}, "1")
	err := usecase.AddClient("1", "Test test", "Test test", "test@test.com", "11980876112", "email", "04417932824")
	if err != nil {
		t.Errorf("TestAddClient failed: %s", err)
	}
	cli, err := usecase.GetClient("1")
	if err != nil {
		t.Errorf("TestAddClient failed: %s", err)
	}
	if cli != "id: 1; name: Test Test; responsible: Test Test; email: test@test.com; phone: +5511980876112; contact: email; document: 044.179.328-24" {
		t.Errorf("TestAddClient failed: %s", cli)
	}
	// terminate
	terminate(usecase)
}

// getUsecase returns a new usecase
func getUsecase(t *testing.T) *Usecase {
	// create a new repository
	dns := os.Getenv("MYSQL_DNS")
	repo, error := repository.NewRepository(dns)
	if error != nil {
		t.Errorf("TestClientOk failed: %s", error)
	}
	// create a new log
	f, err := os.OpenFile("./client_test.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		t.Errorf("TestClientOk failed: %s", error)
	}
	log := log.New(f, "test: ", log.LstdFlags)
	return NewClientUsecase(repo, log)

}

// terminate terminates the usecase
func terminate(usecase *Usecase) {
	usecase.Repo.Close()
	usecase.Log.Writer().(*os.File).Close()
}