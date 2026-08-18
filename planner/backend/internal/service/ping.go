package service

import (
	"planner/internal/dto"
	"planner/internal/port"
)

// PingService is responsible for handling the business logic of the ping endpoint.
type PingService struct {
	logger port.Logger
}

// NewPingService creates a new instance of PingService.
func NewPingService(logger port.Logger) *PingService {
	return &PingService{
		logger: logger,
	}
}

// Run processes the ping request and returns a response indicating the service is alive.
func (s *PingService) Run(request port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Responding to ping request")
	return dto.NewPingResponse(200, "success", "pong")
}
