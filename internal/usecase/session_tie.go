package usecase

import (
	"time"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
	"github.com/lavinas/ephemeris/internal/domain"
)

// SessionTie ties a session to an agenda
func (u *Usecase) SessionTie(dtoIn interface{}) error {
	dtoSessionTie := dtoIn.(*dto.SessionTie)
	if err := dtoSessionTie.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	// get and lock session
	// get and lock agenda
	// make process and massage
	// save and unlock session
	// save and unlock agenda
	// resulta
	return nil
}

// getLockSession gets a session domain for processing and lock it 
func (u *Usecase) getLockSession(id string) (*domain.Session, error) {
	// get and lock session
	return nil, nil
}

// getLockAgenda gets a agenda based on session params and lock if
func (u *Usecase) getLockAgendas(clientId string, serviceId string, At time.Time) ([]*domain.Agenda, error) {
	// get and lock agenda
	return nil, nil
}

// makeProcessAndMessage makes a process and message for the session
func (u *Usecase) makeProcessAndMessage(session *domain.Session, agendas []*domain.Agenda) (string, string) {
	// make process and message
	return "", ""
}






