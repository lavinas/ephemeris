package service

import (
	"planner/internal/port"
)

const (
	logoPath     = "./images/"
	templatePath = "./templates/"
)

// Base is a struct that can be embedded in other service structs to provide common functionality or fields.
type Base struct {
	repo   port.Repository
	logger port.Logger
}

// NewBase creates a new instance of Base with the provided repository and logger.
func NewBase(repo port.Repository, logger port.Logger) *Base {
	return &Base{
		repo:   repo,
		logger: logger,
	}
}
