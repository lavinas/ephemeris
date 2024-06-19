package usecase

import (
	"sort"
	"time"
	"fmt"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionTie ties a session to an agenda
func (u *Usecase) SessionTie(dtoIn interface{}) error {
	dtoSessionTie := dtoIn.(*dto.SessionTie)
	if err := dtoSessionTie.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	session, err := u.getLockSession(dtoIn)
	if err != nil {
		return err
	}
	defer u.unlockSession(session)
	if err := u.untieSession(session); err != nil {
		return err
	}
	if dtoSessionTie.GetCommand() == "tie" {
		if err := u.tieSession(session); err != nil {
			return err
		}
	}
	out := dtoSessionTie.GetOut()
	u.Out = append(u.Out, out.GetDTO(session)...)
	return nil
}

// untieSession unties a session from agendas
func (u *Usecase) untieSession(session *domain.Session) error {
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	if session.AgendaID != "" {
		agenda := domain.Agenda{ID: session.AgendaID}
		if ok, err := agenda.Load(u.Repo); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		} else if !ok {
			return u.error(pkg.ErrPrefBadRequest, pkg.ErrAgendaNotFound, 0, 0)
		}
		agenda.Status = pkg.AgendaStatusOpenned
		if err := u.Repo.Save(&agenda); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
		session.AgendaID = ""
	}
	session.Process = pkg.ProcessStatusOpenned
	if err := u.Repo.Save(session); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}

// tieSession ties session to agendas
func (u *Usecase) tieSession(session *domain.Session) error {
	agendas, err := u.getLockAgenda(session.ClientID, session.ServiceID, session.At)
	if err != nil {
		return err
	}
	defer u.unlockAgendas(agendas)
	if err := u.matchSessionAgendas(session, agendas); err != nil {
		return err
	}
	return nil
}

// matchSessionAgendas matchs session with agenda
func (u *Usecase) matchSessionAgendas(session *domain.Session, agendas []*domain.Agenda) error {
	agenda, err := u.findAgendas(session, agendas)
	if err != nil {
		fmt.Println(1)
		return err
	}
	u.addSessionAgenda(session, agenda)
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	if err := u.Repo.Save(agenda); err != nil {
		return err
	}
	if err := u.Repo.Save(session); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}

// getLockSession gets a session domain for processing and lock it
func (u *Usecase) getLockSession(dtoIn interface{}) (*domain.Session, error) {
	dto := dtoIn.(*dto.SessionTie)
	session := dto.GetDomain()[0].(*domain.Session)
	if err := u.Repo.Begin(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	if ok, err := session.Load(u.Repo); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	} else if !ok {
		return nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrSessionNotFound, 0, 0)
	}
	if session.IsLocked() {
		return nil, u.error(pkg.ErrPrefInternal, pkg.ErrSessionLocked, 0, 0)
	}
	if err := session.Lock(u.Repo); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.Repo.Commit(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return session, nil
}

// getLockAgenda gets a agenda based on session params and lock if
func (u *Usecase) getLockAgenda(clientId string, serviceId string, At time.Time) ([]*domain.Agenda, error) {
	ag := domain.Agenda{ClientID: clientId, ServiceID: serviceId, Status: pkg.AgendaStatusOpenned}
	st1 := At.Add(-time.Hour * 24 * 60)
	st2 := At.Add(time.Hour * 24 * 60)
	if err := u.Repo.Begin(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	agendas, err := ag.LoadRange(u.Repo, st1, st2)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if len(agendas) == 0 {
		return nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrNoAgendasFound, 0, 0)
	}
	for _, a := range agendas {
		if err := a.Lock(u.Repo); err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
	}
	if err := u.Repo.Commit(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return agendas, nil
}

// findAgendas finds agendas linked with session
func (u *Usecase) findAgendas(session *domain.Session, agendas []*domain.Agenda) (*domain.Agenda, error) {
	if len(agendas) == 0 {
		return nil, nil
	}
	agendas = append(agendas, &domain.Agenda{Start: session.At})
	sort.Slice(agendas, func(i, j int) bool {
		return agendas[i].Start.Before(agendas[j].Start)
	})
	idx := -1
	for i := 0; i < len(agendas); i++ {
		if agendas[i].ID == "" {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, u.error(pkg.ErrPrefInternal, pkg.ErrAgendaNotFound, 0, 0)
	}
	agenda := &domain.Agenda{}
	if idx-1 < 0 {
		agenda = agendas[idx+1]
	} else if idx+1 >= len(agendas) {
		agenda = agendas[idx-1]
	} else if session.At.Sub(agendas[idx-1].Start) < agendas[idx+1].Start.Sub(session.At) {
		agenda = agendas[idx-1]
	} else {
		agenda = agendas[idx+1]
	}
	return agenda, nil
}

// saveSessionAgenda saves the session agenda
func (u *Usecase) addSessionAgenda(session *domain.Session, agenda *domain.Agenda) {
	if agenda == nil {
		session.Process = pkg.ProcessStatusUnfound
	} else if session.At.Format("2006-01-02") == agenda.Start.Format("2006-01-02") {
		session.Process = pkg.ProcessStatusLinked
		session.AgendaID = agenda.ID
		agenda.Status = session.Status
	} else {
		session.Process = pkg.ProcessStatusUnconfirmed
		session.AgendaID = agenda.ID
		agenda.Status = pkg.AgendaStatusLocked
	}
}

// unlock session unlocks session
func (u *Usecase) unlockSession(session *domain.Session) error {
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	if err := session.Unlock(u.Repo); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}

// unlock agendas unlocks slice of agendas
func (u *Usecase) unlockAgendas(agendas []*domain.Agenda) error {
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	for _, a := range agendas {
		if err := a.Unlock(u.Repo); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}
