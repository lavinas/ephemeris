package service

import (
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// SessionDelete is a service that handles the deletion of sessions.
type SessionDelete struct {
	repo   port.Repository
	logger port.Logger
}

// NewSessionDelete creates a new instance of SessionDelete with the provided repository.
func NewSessionDelete(repo port.Repository, logger port.Logger) *SessionDelete {
	return &SessionDelete{
		repo:   repo,
		logger: logger,
	}
}

// DeleteSession deletes a session with the given session ID.
func (s *SessionDelete) Run(InDto port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Deleting Session")
	if err := InDto.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation error: %v", err)
		return dto.NewSessionDeleteResponse(400, "Bad Request", err.Error())
	}
	session := &domain.Session{}
	sessionID := InDto.(*dto.SessionDeleteRequest).SessionID
	ret, count, err := session.Find(s.repo, 1, 1, sessionID, "", time.Time{}, time.Time{}, 0, "", "", "")
	if err != nil {
		s.logger.IPrintf(2, "Error checking session existence: %v", err)
		return dto.NewSessionDeleteResponse(500, "Internal Server Error", err.Error())
	}
	if count == 0 {
		return dto.NewSessionDeleteResponse(404, "Not Found", "session not found")
	}
	// Perform soft delete by setting DeletedAt to current time
	now := time.Now()
	ret[0].DeletedAt = &now
	err = s.repo.Save(&ret[0])
	if err != nil {
		s.logger.IPrintf(2, "Error saving session: %v", err)
		return dto.NewSessionDeleteResponse(500, "Internal Server Error", err.Error())
	}
	s.logger.IPrintf(2, "Session with ID %d deleted successfully", sessionID)
	return dto.NewSessionDeleteResponse(200, "OK", "session deleted successfully")
}
