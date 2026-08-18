package service

import (
	"fmt"
	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
	"time"
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
	var session domain.Session
	req := input.(*dto.SessionListRequest)
	dateStart, _ := time.Parse("2006-01-02", req.DateStart)
	dateEnd, _ := time.Parse("2006-01-02", req.DateEnd)
	sessions, err := session.Find(s.repo, req.Page, req.PageSize, req.Nickname, dateStart,
		dateEnd, req.Minutes, req.Status, req.Comments)
	if err != nil {
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
