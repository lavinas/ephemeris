package usecase

import (
	"fmt"
	"time"

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
		a := u.sessionForce(s, &ret)
		if a != nil {
			u.reprocessLinkedSession(s.ID, *a, &ret)
		}
	}
	u.Out = dIn.GetOut().GetDTO(ret)
	return nil
}

// sessionForce links session to a agenda
func (u *Usecase) sessionForce(s *domain.Session, ret *[]interface{}) *string {
	if s.ID == "" ||  s.AgendaID == "" {
		s.Process = fmt.Sprintf("Error: %s", pkg.ErrIdOrAgendaNotFound)
		*ret = append(*ret, s)
		return nil
	
	}
	session, agenda, err := u.getLinkSessionAgenda(s)
	if err != nil {
		s.Process = fmt.Sprintf("Error: %s", err.Error())
		*ret = append(*ret, s)
		return nil
	}
	defer u.unlockSession(session)
	defer u.unlockAgendas([]*domain.Agenda{agenda})
	if err := u.saveLinkedSessionAgenda(session, agenda); err != nil {
		s.Process = fmt.Sprintf("Error: %s", err.Error())
		*ret = append(*ret, s)
		return nil
	}
	*ret = append(*ret, session)
	return &agenda.ID
}

// GetLinkSessionAgenda is a method that returns the session and agenda to be linked
func (u *Usecase) getLinkSessionAgenda(s *domain.Session) (*domain.Session, *domain.Agenda, error) {
	fmt.Println(1, s)
	session, err := u.getLockSession(s.ID)
	if err != nil {
		fmt.Println(2, s.ID, err.Error())
		return nil, nil, err
	}
	agendas, err := u.getLockAgenda(&domain.Agenda{ID: s.AgendaID}, time.Time{}, time.Time{}, nil)
	if err != nil {
		return nil, nil, err
	}
	if len(agendas) == 0 {
		return nil, nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrAgendaNotFound, 0, 0)
	}
	if len(agendas) > 1 {
		return nil, nil, u.error(pkg.ErrPrefInternal, pkg.ErrAgendaMultiple, 0, 0)
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
func (u *Usecase) reprocessLinkedSession(sessionID string, agendaID string, ret *[]interface{}) {
	tx := u.Repo.Begin()
	defer u.Repo.Rollback(tx)
	add := fmt.Sprintf("id != '%s'", sessionID)
	sl, _, err := u.Repo.Find(tx, &domain.Session{AgendaID: agendaID}, -1, add)
	if err != nil || sl == nil {
		return
	}
	sessions := *sl.(*[]domain.Session)
	session := sessions[0]
	session.Process = pkg.ProcessStatusOpenned
	session.AgendaID = ""
	if err := u.tieCommand(&session); err != nil {
		session.Process = fmt.Sprintf("Error: %s", err.Error())
		*ret = append(*ret, &session)
		return
	}
	*ret = append(*ret, &session)
}
