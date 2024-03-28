package usecase

import (
	"reflect"
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
	add := &dto.ClientAdd{Object: "client", Action: "add", ID: "1", Name: "Test Test",
		Responsible: "Test Test", Email: "test@test.com", Phone: "11980876112", Contact: "email", Document: "04417932824"}
	if _,_, err := usecase.ClientAdd(add); err != nil {
		t.Errorf("TestAddClient failed: %s", err)
	}
	in := &dto.ClientGet{Object: "client", Action: "get", ID: "1"}
	resp := &dto.ClientGet{Object: "client", Action: "get", ID: "1", Name: "Test Test", Responsible: "Test Test",
		Email: "test@test.com", Phone: "+5511980876112", Contact: "email", Document: "044.179.328-24"}
	iout, _, err := usecase.ClientGet(in)
	if err != nil {
		t.Errorf("TestGetClient failed: %s", err)
	}
	out := iout.([]dto.ClientGet)
	if len(out) != 1 {
		t.Errorf("TestGetClient failed: expected 1, got %d", len(out))
	}
	if !reflect.DeepEqual(out[0], resp) {
		t.Errorf("TestGetClient failed: expected %v, got %v", resp, out[0])
	}
	// terminate
	terminate(usecase)
}
