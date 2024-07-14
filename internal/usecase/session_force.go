package usecase

import (
	"time"
	"fmt"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionForce links sessions to agendas
func (u *Usecase) SessionForce(dtoIn interface{}) error {
	dIn, _ := dtoIn.(*dto.SessionForce)
	if err := dIn.Validate(); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	domains := dIn.GetDomain()
	ret := []interface{}{}
	for _, d := range domains {
		s := d.(*domain.Session)
		u.sessionForce(s, &ret)
		u.reprocessLinkedSession(s.ID, &ret)
	}
	u.Out = dIn.GetOut().GetDTO(ret)
	return nil
}

// sessionForce links session to a agenda
func (u *Usecase) sessionForce(s *domain.Session, ret *[]interface{}) {
	session, agenda, err := u.getLinkSessionAgenda(s)
	if err != nil {
		s.Process = fmt.Sprintf("Error: %s", err.Error())
		*ret = append(*ret, s)
		return 
	}
	defer u.unlockSession(session)
	defer u.unlockAgendas([]*domain.Agenda{agenda})
	if err := u.saveLinkedSessionAgenda(session, agenda); err != nil {
		s.Process = fmt.Sprintf("Error: %s", err.Error())
		*ret = append(*ret, s)
		return 
	}
	*ret = append(*ret, session)
}

// GetLinkSessionAgenda is a method that returns the session and agenda to be linked
func (u *Usecase) getLinkSessionAgenda(s *domain.Session) (*domain.Session, *domain.Agenda, error) {
	session, err := u.getLockSession(s.ID)
	if err != nil {
		return nil, nil, err
	}
	agendas, err := u.getLockAgenda(&domain.Agenda{ID: s.AgendaID}, time.Time{}, time.Time{}, nil)
	if err != nil {
		return nil, nil, err
	}
	if len(agendas) == 0 {
		return nil, nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrAgendaNotFound, 0, 0)
	}
	return session, agendas[0], nil
}

// SaveLinkedSessionAgenda is a method that saves the linked session and agenda
func (u *Usecase) saveLinkedSessionAgenda(session *domain.Session, agenda *domain.Agenda) error {
	if session.ClientID != agenda.ClientID {
		return u.error(pkg.ErrPrefBadRequest, pkg.ErrAgendaClientMismatch, 0, 0)
	}
	session.AgendaID = agenda.ID
	session.Process = pkg.ProcessStatusLinked
	agenda.Status = session.Status
	if err := u.saveSessionAgenda(session, agenda); err != nil {
		return err
	}
	return nil
}

// reprocessLinkedSession is a method that reprocesses the linked session
func (u *Usecase) reprocessLinkedSession(sessionID string, ret *[]interface{}) {
	tx := u.Repo.Begin()
	defer u.Repo.Rollback(tx)
	sl, _, err := u.Repo.Find(tx, &domain.Session{AgendaID: sessionID}, 0)
	if err != nil || sl == nil {
		return
	}
	s := sl.([]*domain.Session)[0]
	if s.ID == sessionID {
		return
	}
	if err := u.tieCommand(s); err != nil {
		s.Process = fmt.Sprintf("Error: %s", err.Error())
		*ret = append(*ret, s)
		return
	}
	*ret = append(*ret, s)
}
