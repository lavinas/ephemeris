package usecase

import (
	"log"
	"testing"
	"os"
	"strings"

	"github.com/lavinas/ephemeris/internal/adapter/repository"

)

func Test1 (t *testing.T) {
	x := ""
	y := strings.Split(x, " ")
	if len(y) < 2 {
		t.Errorf("Test1 failed: %s", y)
	}
}

func TestCommandClient(t *testing.T) {
	// prepare
	usecase := getUsecase(t)
	// test valid
	cmd := "client add name Paulo Barbosa"
	result := usecase.Command(cmd)
	if result != "ok" {
		t.Errorf("TestCommandClient failed: %s", result)
	}
	// test invalid command
	cmd = "ipsus add name Paulo Barbosa"
	result = usecase.Command(cmd)
	if result != "command not found: ipsus, possible commands: client" {
		t.Errorf("TestCommandClient failed: %s", result)
	}
	// test short command
	cmd = "client"
	result = usecase.Command(cmd)
	if result != "command should have at least 2 words" {
		t.Errorf("TestCommandClient failed: %s", result)
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
