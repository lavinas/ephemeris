package service

import (
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockSessionUpdateRepo struct {
	mockPlanRepoForService
	savedSession *domain.Session
}

func (m *mockSessionUpdateRepo) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	desc := "online/canto/30"
	s := domain.Session{
		ID:             1,
		CustomerID:     1,
		ServiceID:      1,
		Customer:       &domain.Customer{ID: 1, Nickname: "cliente_1"},
		Service:        &domain.Service{ID: 1, Name: "online/canto/30", Description: &desc},
		SessionDate:    time.Now(),
		SessionMinutes: 30,
		SessionStatus:  "realizada",
	}
	return []interface{}{s}, nil
}

func (m *mockSessionUpdateRepo) FindCount(conditions map[string]interface{}) (int64, error) {
	return 1, nil
}

func (m *mockSessionUpdateRepo) Save(model interface{}) error {
	if sess, ok := model.(*domain.Session); ok {
		m.savedSession = sess
	}
	return nil
}

func (m *mockSessionUpdateRepo) GetServiceByID(id int64) (*domain.Service, error) {
	return &domain.Service{ID: id, Name: "online/canto/60"}, nil
}

func TestSessionUpdate_ChangeService(t *testing.T) {
	mockRepo := &mockSessionUpdateRepo{
		mockPlanRepoForService: *newMockPlanRepoForService(),
	}
	logger := &mockLogger{}
	svc := NewSessionUpdate(mockRepo, logger)

	req := &dto.SessionUpdateRequest{
		ID:        1,
		ServiceID: 3,
		Minutes:   60,
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("Expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if mockRepo.savedSession == nil {
		t.Fatal("Expected savedSession to be populated")
	}

	if mockRepo.savedSession.ServiceID != 3 {
		t.Errorf("Expected ServiceID 3, got %d", mockRepo.savedSession.ServiceID)
	}

	if mockRepo.savedSession.Service != nil {
		t.Errorf("Expected preloaded Service association to be nil to prevent GORM reversion, got %+v", mockRepo.savedSession.Service)
	}

	if mockRepo.savedSession.Customer != nil {
		t.Errorf("Expected preloaded Customer association to be nil to prevent GORM reversion, got %+v", mockRepo.savedSession.Customer)
	}
}
