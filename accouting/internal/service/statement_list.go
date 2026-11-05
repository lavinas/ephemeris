package service

import (
	"fmt"
	"math"
	"time"

	"accounting/internal/dto"
	"accounting/internal/port"
)

// StatementListService handles statement listing with pagination and filters
type StatementListService struct {
	repo   port.Repository
	logger port.Logger
}

// NewStatementListService creates a new StatementListService
func NewStatementListService(repo port.Repository, logger port.Logger) *StatementListService {
	return &StatementListService{
		repo:   repo,
		logger: logger,
	}
}

// Run executes the list query
func (s *StatementListService) Run(req *dto.StatementListRequest) *dto.StatementListResponse {
	if err := req.Validate(s.repo); err != nil {
		return &dto.StatementListResponse{
			ResponseBase: dto.NewResponseBase(400, "error", err.Error()),
		}
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		}
	}

	offset := (req.Page - 1) * req.PageSize
	items, total, err := s.repo.ListStatements(req.AccountID, startDate, endDate, req.Type, offset, req.PageSize)
	if err != nil {
		s.logger.IPrintf(2, "Error listing statements: %v", err)
		return &dto.StatementListResponse{
			ResponseBase: dto.NewResponseBase(500, "error", fmt.Sprintf("failed to list statements: %v", err)),
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	var data []dto.StatementItemResponse
	for _, item := range items {
		data = append(data, dto.ToStatementItemResponse(item))
	}

	resp := &dto.StatementListResponse{
		ResponseBase: dto.NewResponseBase(200, "success", "Statements retrieved successfully"),
		AccountID:    req.AccountID,
		Total:        total,
		Page:         req.Page,
		PageSize:     req.PageSize,
		TotalPages:   totalPages,
		Data:         data,
	}
	return resp
}

