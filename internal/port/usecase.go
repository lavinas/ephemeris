package port

import (
	"github.com/lavinas/ephemeris/internal/dto"
)

// UseCase is an interface that defines the methods for the use case
type UseCase interface {
	Command(string) string
	AddClient(*dto.ClientAdd) error
	GetClient(*dto.ClientGet) error
}
