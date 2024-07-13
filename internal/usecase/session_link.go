package usecase

import (
	"time"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionLink links sessions to agendas
func (u *Usecase) SessionLink(dtoIn *dto.SessionLink) error {
	if err := dtoIn.Validate(); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	domains := dtoIn.GetDomain()
	for _, d := range domains {
		s := d.(*domain.Session)
		session, agenda, err := u.getLinkSessionAgenda(s)
		if err != nil {
			return err
		}
		if err := u.saveLinkedSessionAgenda(session, agenda); err != nil {
			return err
		}
		if err := u.reprocessLinkedSession(session.ID); err != nil {
			return err
		}
	}
	return nil
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
	session.AgendaID = agenda.ID
	session.Process = pkg.ProcessStatusLinked
	agenda.Status = session.Status
	if err:= u.saveSessionAgenda(session, agenda); err != nil {
		return err
	}
	return nil
}

// reprocessLinkedSession is a method that reprocesses the linked session
func (u *Usecase) reprocessLinkedSession(sessionID string) error {
	tx := u.Repo.Begin()
	defer u.Repo.Rollback(tx)
	sl, _, err := u.Repo.Find(tx, &domain.Session{AgendaID: sessionID}, 0)
	if err != nil {
		return err
	}
	s2 := sl.([]*domain.Session)
	if err := u.tieCommand(s2[0]); err != nil {
		return err
	}
	return nil
}
