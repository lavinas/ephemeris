package usecase

import (
	"testing"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
)

func TestClient(t *testing.T) {
	// prepare
	usecase := getUsecase(t)
	// add a client ok
	usecase.Repo.Delete(&domain.Client{}, "1")
	defer usecase.Repo.Delete(&domain.Client{}, "1")
	dto := &dto.ClientAdd{
		ID: "1", Name: "Test test", Responsible: "Test test", Email: "test@test.com",
		Phone: "11980876112", Contact: "email", Document: "04417932824",
	}
	err := usecase.AddClient(dto)
	if err != nil {
		t.Errorf("TestAddClient failed: %s", err)
	}
	cli, err := usecase.GetClient("1")
	if err != nil {
		t.Errorf("TestAddClient failed: %s", err)
	}
	if cli.ID != "1" {
		t.Errorf("TestAddClient ID failed: %s", cli.ID)
	}
	if cli.Name != "Test Test" {
		t.Errorf("TestAddClient Name failed: %s", cli.Name)
	}
	if cli.Responsible != "Test Test" {
		t.Errorf("TestAddClient Responsible failed: %s", cli.Responsible)
	}
	if cli.Email != "test@test.com" {
		t.Errorf("TestAddClient Email failed: %s", cli.Email)
	}
	if cli.Phone != "+5511980876112" {
		t.Errorf("TestAddClient Phone failed: %s", cli.Phone)
	}
	if cli.Contact != "email" {
		t.Errorf("TestAddClient Contact failed: %s", cli.Contact)
	}
	if cli.Document != "044.179.328-24" {
		t.Errorf("TestAddClient Document failed: %s", cli.Document)
	}
	// terminate
	terminate(usecase)
}

