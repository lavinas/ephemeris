package service

import (
	"fmt"

	"accounting/internal/domain"
	"accounting/internal/dto"
	"accounting/internal/port"
)

// AccountActionService manages CRUD operations for bank account actions and rules
type AccountActionService struct {
	repo   port.Repository
	logger port.Logger
}

func NewAccountActionService(repo port.Repository, logger port.Logger) *AccountActionService {
	return &AccountActionService{
		repo:   repo,
		logger: logger,
	}
}

// ListByAccount returns all actions with their rules for an account
func (s *AccountActionService) ListByAccount(accountID int64) *dto.AccountActionListResponse {
	if accountID <= 0 {
		return &dto.AccountActionListResponse{
			ResponseBase: dto.NewResponseBase(400, "error", "account_id é obrigatório"),
		}
	}

	actions, err := s.repo.GetAccountActions(accountID)
	if err != nil {
		s.logger.IPrintf(2, "Error fetching account actions for account %d: %v", accountID, err)
		return &dto.AccountActionListResponse{
			ResponseBase: dto.NewResponseBase(500, "error", fmt.Sprintf("failed to get account actions: %v", err)),
		}
	}

	data := make([]dto.AccountActionItemDTO, len(actions))
	for i, a := range actions {
		data[i] = dto.ToAccountActionItemDTO(a)
	}

	return &dto.AccountActionListResponse{
		ResponseBase: dto.NewResponseBase(200, "success", "Ações recuperadas com sucesso"),
		AccountID:    accountID,
		Data:         data,
	}
}

// GetByID returns an action by its ID
func (s *AccountActionService) GetByID(id int64) *dto.AccountActionDetailResponse {
	if id <= 0 {
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(400, "error", "id é obrigatório"),
		}
	}

	action, err := s.repo.GetActionByID(id)
	if err != nil {
		s.logger.IPrintf(2, "Error fetching action %d: %v", id, err)
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(500, "error", fmt.Sprintf("failed to get action: %v", err)),
		}
	}
	if action == nil {
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(404, "error", "Ação não encontrada"),
		}
	}

	item := dto.ToAccountActionItemDTO(*action)
	return &dto.AccountActionDetailResponse{
		ResponseBase: dto.NewResponseBase(200, "success", "Ação recuperada com sucesso"),
		Data:         &item,
	}
}

// Create inserts a new action with rules
func (s *AccountActionService) Create(req *dto.AccountActionCreateRequest) *dto.AccountActionDetailResponse {
	if err := req.Validate(); err != nil {
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(400, "error", err.Error()),
		}
	}

	action := &domain.BankAccountAction{
		AccountID: req.AccountID,
		Action:    req.Action,
	}
	for _, r := range req.Rules {
		action.Rules = append(action.Rules, domain.BankAccountActionRule{
			Field:         r.Field,
			Operator:      r.Operator,
			Value:         r.Value,
			CaseSensitive: r.CaseSensitive,
		})
	}

	if err := s.repo.CreateAccountAction(action); err != nil {
		s.logger.IPrintf(2, "Error creating account action: %v", err)
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(500, "error", fmt.Sprintf("failed to create action: %v", err)),
		}
	}

	created, _ := s.repo.GetActionByID(action.ID)
	if created == nil {
		created = action
	}
	item := dto.ToAccountActionItemDTO(*created)

	return &dto.AccountActionDetailResponse{
		ResponseBase: dto.NewResponseBase(201, "success", "Ação cadastrada com sucesso"),
		Data:         &item,
	}
}

// Update updates an existing action and replaces its rules
func (s *AccountActionService) Update(req *dto.AccountActionUpdateRequest) *dto.AccountActionDetailResponse {
	if err := req.Validate(); err != nil {
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(400, "error", err.Error()),
		}
	}

	existing, err := s.repo.GetActionByID(req.ID)
	if err != nil {
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(500, "error", fmt.Sprintf("failed to get existing action: %v", err)),
		}
	}
	if existing == nil {
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(404, "error", "Ação não encontrada"),
		}
	}

	action := &domain.BankAccountAction{
		ID:        req.ID,
		AccountID: req.AccountID,
		Action:    req.Action,
	}
	for _, r := range req.Rules {
		action.Rules = append(action.Rules, domain.BankAccountActionRule{
			ActionID:      req.ID,
			Field:         r.Field,
			Operator:      r.Operator,
			Value:         r.Value,
			CaseSensitive: r.CaseSensitive,
		})
	}

	if err := s.repo.UpdateAccountAction(action); err != nil {
		s.logger.IPrintf(2, "Error updating action %d: %v", req.ID, err)
		return &dto.AccountActionDetailResponse{
			ResponseBase: dto.NewResponseBase(500, "error", fmt.Sprintf("failed to update action: %v", err)),
		}
	}

	updated, _ := s.repo.GetActionByID(action.ID)
	if updated == nil {
		updated = action
	}
	item := dto.ToAccountActionItemDTO(*updated)

	return &dto.AccountActionDetailResponse{
		ResponseBase: dto.NewResponseBase(200, "success", "Ação atualizada com sucesso"),
		Data:         &item,
	}
}

// Delete soft deletes an action
func (s *AccountActionService) Delete(id int64) *dto.ResponseBase {
	if id <= 0 {
		resp := dto.NewResponseBase(400, "error", "id é obrigatório")
		return &resp
	}

	if err := s.repo.DeleteAccountAction(id); err != nil {
		s.logger.IPrintf(2, "Error deleting action %d: %v", id, err)
		resp := dto.NewResponseBase(500, "error", fmt.Sprintf("failed to delete action: %v", err))
		return &resp
	}

	resp := dto.NewResponseBase(200, "success", "Ação excluída com sucesso")
	return &resp
}

