package usecase

import (
	"log"
	"os"
	"testing"

	"github.com/lavinas/ephemeris/internal/adapters/repository"
)

// getUsecase returns a new usecase
func getUsecase(t *testing.T) *Usecase {
	// create a new repository
	dns := os.Getenv("MYSQL_DNS")
	repo, error := repository.NewRepository(dns)
	if error != nil {
		t.Errorf("TestClientOk failed: %s", error)
	}
	// create a new log
	f, err := os.OpenFile("/dev/stdout", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Errorf("TestClientOk failed: %s", error)
	}
	log := log.New(f, "test: ", log.LstdFlags)
	return NewUsecase(repo, log)

}

// terminate terminates the usecase
func terminate(usecase *Usecase) {
	usecase.Repo.Close()
	usecase.Log.Writer().(*os.File).Close()
}
