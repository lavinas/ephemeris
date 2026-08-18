package service

import (
	"fmt"
	"planner/internal/dto"
	"planner/internal/port"
	"planner/internal/domain"
)

// SessionList is a service that handles the retrieval of session lists.
type SessionList struct {
	Base
}

// NewSessionList creates a new instance of SessionList with the provided repository and logger.
func NewSessionList(repo port.Repository, logger port.Logger) *SessionList {
	return &SessionList{
		Base: *NewBase(repo, logger),
	}
}

// Run executes the session list retrieval process using the provided input DTO and returns an output DTO.
func (s *SessionList) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing list sessions request: %v", input)
	// Validate the input DTO
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewSessionListResponse(400, "error", err.Error(), nil)
	}
	// Retrieve sessions from the repository based on the input DTO
	sessions, err := s.findSessions(input)
	if err != nil {
		s.logger.IPrintf(2, "Failed to retrieve sessions: %v", err)
		return dto.NewSessionListResponse(500, "error", err.Error(), nil)
	}
	// Create and return the output DTO with the retrieved sessions
	s.logger.IPrintf(2, "Successfully retrieved %d sessions", len(sessions))
	message := fmt.Sprintf("retrieve %d registers", len(sessions))
	return dto.NewSessionListResponse(200, "success", message, s.dtoOut(sessions))
}

// findSessions is a helper function to find sessions in the repository based on the input DTO.
func (s *SessionList) findSessions(input port.InDTO) ([]domain.Session, error) {
	// Assuming input is of type SessionListRequest
	req := input.(*dto.SessionListRequest)
	conditions := []interface{}{}
	if req.Nickname != "" {
		conditions = append(conditions, "nickname = ?", req.Nickname)
	}
	if req.DateStart != "" {
		conditions = append(conditions, "date >=", req.DateStart, req.DateEnd)
	}
	if req.DateEnd != "" {
		conditions = append(conditions, "date <=", req.DateEnd)
	}
	if req.Minutes != 0 {
		conditions = append(conditions, "minutes = ?", req.Minutes)
	}
	if req.Status != "" {
		conditions = append(conditions, "status = ?", req.Status)
	}
	if req.Comments != "" {
		conditions = append(conditions, "comments LIKE ?", "%"+req.Comments+"%")
	}
	var sessions []domain.Session

	if err := s.repo.Find(&sessions, req.Page, req.PageSize, conditions...); err != nil {
		return nil, err
	}
	return sessions, nil
}

// dtoOut is a helper function to convert a slice of Session models to a slice of Session DTOs.
func (s *SessionList) dtoOut(sessions []domain.Session) []dto.Session {
	dtoSessions := make([]dto.Session, len(sessions))
	for i, session := range sessions {
		comments := ""
		if session.Comments != nil {
			comments = *session.Comments
		}
		dtoSessions[i] = dto.Session{
			SessionID: session.ID,
			Nickname:  session.CustomerNickname,
			Date:      session.SessionDate.Format("2006-01-02"),
			Minutes:   session.SessionMinutes,
			Status:    session.SessionStatus,
			Comments:  comments,
		}
	}
	return dtoSessions
}
