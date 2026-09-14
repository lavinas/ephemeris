package service

import (
	"fmt"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// SessionUsers is a service that handles the retrieval of session users.
type SessionUsers struct {
	Base
}

// NewSessionUsers creates a new instance of SessionUsers with the provided repository and logger.
func NewSessionUsers(repo port.Repository, logger port.Logger) *SessionUsers {
	return &SessionUsers{
		Base: *NewBase(repo, logger),
	}
}

// Run executes the session users retrieval process using the provided input DTO and returns an output DTO.
func (s *SessionUsers) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing session users request: %v", input)

	dtoIn, ok := input.(*dto.SessionUsersRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input DTO type: %T", input)
		return dto.NewSessionUsersResponse(400, "error", "Invalid input DTO type", nil)
	}
	if err := dtoIn.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Invalid session users request: %v", err)
		return dto.NewSessionUsersResponse(400, "error", err.Error(), nil)
	}
	users, err := s.getUsers(*dtoIn)
	if err != nil {
		s.logger.IPrintf(2, "Error retrieving session users: %v", err)
		return dto.NewSessionUsersResponse(500, "error", err.Error(), nil)
	}
	// Create and return the output DTO with the retrieved session users
	s.logger.IPrintf(2, "Successfully retrieved %d session users", len(users))
	message := fmt.Sprintf("Successfully retrieved %d session users", len(users))
	dtoOut := dto.NewSessionUsersResponse(200, "success", message, users)
	s.logger.IPrintf(2, "Returning response: %v", dtoOut)
	return dtoOut
}

// getUsers is a helper function that retrieves session users based on the input DTO.
func (s *SessionUsers) getUsers(input dto.SessionUsersRequest) ([]string, error) {
	sd, _ := time.Parse("2006-01-02", input.StartDate)
	ed, _ := time.Parse("2006-01-02", input.EndDate)

	session := domain.Session{}
	users, err := session.FindUsers(
		s.repo,
		sd,
		ed,
		input.Minutes,
		input.Service,
		input.Status,
	)
	if err != nil {
		return nil, err
	}
	return users, nil
}
