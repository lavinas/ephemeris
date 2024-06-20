package usecase

import (
	"sort"
	"time"

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
	agenda, err := u.restartLockAgenda(session.AgendaID)
	if err != nil {
		return err
	}
	if agenda != nil {
		defer u.unlockAgendas([]*domain.Agenda{agenda})
	}
	session.Process = pkg.ProcessStatusOpenned
	session.AgendaID = ""
	if err := u.saveSessionAgenda(session, agenda); err != nil {
		return err
	}
	return nil
}

// restartAgenda restarts agenda status
func (u *Usecase) restartLockAgenda(id string) (*domain.Agenda, error) {
	if id == "" {
		return nil, nil
	}
	agenda := &domain.Agenda{ID: id}
	agendas, err := u.getLockAgenda(agenda, time.Time{}, time.Time{})
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	agenda = agendas[0]
	agenda.Status = pkg.AgendaStatusOpenned
	return agenda, nil
}

// tieSession ties session to agendas
func (u *Usecase) tieSession(session *domain.Session) error {
	ag := domain.Agenda{ClientID: session.ClientID, ServiceID: session.ServiceID}
	agendas, err := u.getLockAgenda(&ag, session.At.Add(-time.Hour*24*60), session.At.Add(time.Hour*24*60))
	if err != nil {
		return err
	}
	defer u.unlockAgendas(agendas)
	agenda, err := u.findAgenda(session, agendas)
	if err != nil {
		return err
	}
	u.matchSessionAgenda(session, agenda)
	if err := u.saveSessionAgenda(session, agenda); err != nil {
		return err
	}
	return nil
}

// saveSessionAgenda saves the session agenda
func (u *Usecase) saveSessionAgenda(session *domain.Session, agenda *domain.Agenda) error {
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	if agenda != nil {
		if err := u.Repo.Save(agenda); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
	}
	if session != nil {
		if err := u.Repo.Save(session); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
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
func (u *Usecase) getLockAgenda(agenda *domain.Agenda, start time.Time, end time.Time) ([]*domain.Agenda, error) {
	if err := u.Repo.Begin(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	agendas, err := agenda.LoadRange(u.Repo, start, end)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if len(agendas) == 0 {
		return nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrNoAgendasFound, 0, 0)
	}
	for _, a := range agendas {
		if a.IsLocked() {
			return nil, u.error(pkg.ErrPrefInternal, pkg.ErrAgendaLocked, 0, 0)
		}
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
func (u *Usecase) findAgenda(session *domain.Session, agendas []*domain.Agenda) (*domain.Agenda, error) {
	if session == nil || len(agendas) == 0 {
		return nil, nil
	}
	ags, idx := u.getOrderedAgendas(session, agendas)
	agenda := &domain.Agenda{}
	switch {
	case idx == -1:
		return nil, u.error(pkg.ErrPrefInternal, pkg.ErrAgendaNotFound, 0, 0)
	case idx-1 < 0:
		agenda = ags[idx+1]
	case idx+1 >= len(ags):
		agenda = ags[idx-1]
	case session.At.Sub(ags[idx-1].Start) < ags[idx+1].Start.Sub(session.At):
		agenda = ags[idx-1]
	default:
		agenda = ags[idx+1]
	}
	return agenda, nil
}

// getorederedAgendas gets ordered agendas
func (u *Usecase) getOrderedAgendas(session *domain.Session, agendas []*domain.Agenda) ([]*domain.Agenda, int) {
	ags := []*domain.Agenda{}
	ags = append(ags, agendas...)
	ags = append(ags, &domain.Agenda{ID: "**", Start: session.At})
	sort.Slice(ags, func(i, j int) bool {
		return ags[i].Start.Before(ags[j].Start)
	})
	idx := -1
	for i, a := range ags {
		if a.ID == "**" {
			idx = i
			break
		}
	}
	return ags, idx
}

// saveSessionAgenda saves the session agenda
func (u *Usecase) matchSessionAgenda(session *domain.Session, agenda *domain.Agenda) {
	switch {
	case agenda == nil:
		session.Process = pkg.ProcessStatusUnfound
	case session.At.Format("2006-01-02") == agenda.Start.Format("2006-01-02"):
		session.Process = pkg.ProcessStatusLinked
		session.AgendaID = agenda.ID
		agenda.Status = session.Status
	default:
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
