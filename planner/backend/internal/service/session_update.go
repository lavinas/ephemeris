package service

import (
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// SessionUpdate is a service that handles the updating of sessions.
type SessionUpdate struct {
	repo   port.Repository
	logger port.Logger
}

// NewSessionUpdate creates a new instance of SessionUpdate with the provided repository.
func NewSessionUpdate(repo port.Repository, logger port.Logger) *SessionUpdate {
	return &SessionUpdate{
		repo:   repo,
		logger: logger,
	}
}

// UpdateSession updates an existing session with the provided data.
func (s *SessionUpdate) Run(InDto port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Updating Session")
	req := InDto.(*dto.SessionUpdateRequest)
	// Validate the request data
	if err := req.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation error: %v", err)
		return dto.NewSessionUpdateResponse(400, "Bad Request", err.Error())
	}
	// Get the session model from the request
	sessionModel := req.GetSession()
	// Update the session model with the new data from the request
	if err := s.updateSessionModel(&sessionModel, req); err != nil {
		s.logger.IPrintf(2, "Failed to update session model: %v", err)
		return dto.NewSessionUpdateResponse(500, "error", err.Error())
	}
	s.logger.IPrintf(2, "Session model updated: %+v", sessionModel)

	// Save the session model to the repository
	if err := s.repo.Save(&sessionModel); err != nil {
		s.logger.IPrintf(2, "Failed to save session: %v", err)
		return dto.NewSessionUpdateResponse(500, "error", err.Error())
	}
	// Create and return the output DTO with the session ID
	s.logger.IPrintf(2, "Successfully updated session with ID: %d", sessionModel.ID)
	return dto.NewSessionUpdateResponse(200, "success", "session updated")
}

// updateSessionModel updates the session model with the data from the request.
func (s *SessionUpdate) updateSessionModel(sessionModel *domain.Session, req *dto.SessionUpdateRequest) error {
	if req.Nickname != "" {
		sessionModel.CustomerNickname = req.Nickname
	}
	if req.Date != "" {
		// Parse the date string into a time.Time object
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return err
		}
		sessionModel.SessionDate = parsedDate
	}
	if req.Minutes > 0 {
		sessionModel.SessionMinutes = req.Minutes
	}
	if req.Service != "" {
		sessionModel.SessionService = req.Service
	}
	if req.Status != "" {
		sessionModel.SessionStatus = req.Status
	}
	if req.Comments != "" {
		sessionModel.Comments = &req.Comments
	} else {
		sessionModel.Comments = nil
	}
	return nil
}
