package service

import (
	"context"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// SessionCreate is  a service that handles the creation of sessions.
type SessionCreate struct {
	Base
	comptime port.ComptimeService
}

// NewSessionCreate
func NewSessionCreate(repo port.Repository, logger port.Logger) *SessionCreate {
	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = NewComptime(cr, logger)
	}
	return &SessionCreate{
		Base:     *NewBase(repo, logger),
		comptime: comptime,
	}
}

// SetComptimeService sets the comptime service.
func (s *SessionCreate) SetComptimeService(cs port.ComptimeService) {
	s.comptime = cs
}


// Run executes the session creation process using the provided input DTO and returns an output DTO.
func (s *SessionCreate) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create session request: %v", input)
	// Validate the input DTO
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewSessionCreateResponse(400, "error", err.Error(), 0)
	}
	// Create a new session model from the input DTO
	sessionModel := s.createSessionModel(input)

	// Save the session model to the repository
	if err := s.repo.Save(sessionModel); err != nil {
		s.logger.IPrintf(2, "Failed to save session: %v", err)
		return dto.NewSessionCreateResponse(500, "error", err.Error(), 0)
	}

	sessionID := sessionModel.(*domain.Session).ID

	// Trigger comptime sync
	if s.comptime != nil {
		if err := s.comptime.ProcessSessionCreated(context.Background(), sessionID); err != nil {
			s.logger.IPrintf(2, "Aviso: falha na sincronização do banco de horas da sessão %d: %v", sessionID, err)
		}
	}

	// Create and return the output DTO with the session ID
	s.logger.IPrintf(2, "Successfully created session with ID: %d", sessionID)
	return dto.NewSessionCreateResponse(200, "success", "session created", sessionID)
}

// createSessionModel creates a session model from the input DTO.
func (s *SessionCreate) createSessionModel(input port.InDTO) interface{} {
	// Assuming input is of type SessionCreateRequest
	req := input.(*dto.SessionCreateRequest)
	dtime, _ := time.Parse("2006-01-02", req.Date)
	var comments *string
	if req.Comments != "" {
		comments = &req.Comments
	}
	return domain.NewSession(req.CustomerID, req.ServiceID, dtime, req.Minutes, req.Status, comments)
}
