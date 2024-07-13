package usecase

import (
	"time"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/pkg"


)


// SessionLink links sessions to agendas
func (u *Usecase) SessionLink(s *dto.SessionLink) error {
	if err := s.Validate(); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	domains := s.GetDomain()
	for _, d := range domains {
		s := d.(*domain.Session)
		session, err := u.getLockSession(s.ID)
		if err != nil {
			return err
		}
		agenda := &domain.Agenda{ID: s.AgendaID}
		agendas, err := u.getLockAgenda(agenda, time.Time{}, time.Time{}, nil)
		if err != nil {
			return err
		}
		if len(agendas) == 0 {
			return u.error(pkg.ErrPrefBadRequest, pkg.ErrAgendaNotFound, 0, 0)
		}
		tx := u.Repo.Begin()
		defer u.Repo.Rollback(tx)
		sl, _, err := u.Repo.Find(tx, &domain.Session{AgendaID: s.AgendaID}, 0)
		if err != nil {
			s2 := sl.([]*domain.Session)
			if err := u.tieCommand(s2[0]); err != nil {
				return err
			}
		}
		agenda = agendas[0]
		session.AgendaID = agenda.ID
		session.Process = pkg.ProcessStatusLinked
		agenda.Status = session.Status
		if err:= u.saveSessionAgenda(session, agenda); err != nil {
			return err
		}
	}
	return nil
}
