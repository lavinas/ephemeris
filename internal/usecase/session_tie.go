package usecase

import (
	"slices"
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
	sessions, err := u.findSessionsTie(dtoSessionTie)
	if err != nil {
		return err
	}
	command := dtoSessionTie.GetCommand()
	result := []interface{}{}
	for _, session := range *sessions {
		s, err := u.sessionTieOne(session.ID, command)
		if err != nil {
			session.Process = pkg.ProcessStatusError
			session.AgendaID = err.Error()
			result = append(result, &session)
			continue
		}
		result = append(result, s)
	}
	if len(result) == 0 {
		return u.error(pkg.ErrPrefBadRequest, pkg.ErrNoSessionsProcessed, 0, 0)
	}
	out := dtoSessionTie.GetOut()
	u.Out = out.GetDTO(result)
	return nil
}

// sessionSortFunc is a function to sort sessions
func sessionSortFunc(a domain.Session, b domain.Session) int {
	switch {
	case a.At.Before(b.At):
		return -1
	case a.At.After(b.At):
		return 1
	default:
		return 0
	}
}

// findSessionsTie finds a session to tie
func (u *Usecase) findSessionsTie(dtoIn *dto.SessionTie) (*[]domain.Session, error) {
	d, extras, err := dtoIn.GetInstructions(dtoIn.GetDomain()[0])
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.Repo.Begin(""); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback("")
	if err := d.Format(u.Repo, "filled", "noduplicity"); err != nil {
		return nil, u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	base, _, err := u.Repo.Find(d, -1, "", extras...)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if base == nil {
		return nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrSessionNotFound, 0, 0)
	}
	ret := base.(*[]domain.Session)
	slices.SortFunc(*ret, sessionSortFunc)
	return ret, nil
}

// sessionTieOne ties a session to an agenda
func (u *Usecase) sessionTieOne(id string, command string) (*domain.Session, error) {
	session, err := u.getLockSession(id)
	if err != nil {
		return nil, err
	}
	defer u.unlockSession(session)
	if err := u.untieSession(session); err != nil {
		return nil, err
	}
	if command == "tie" {
		s := session
		for s != nil {
			if s, err = u.tieSession(s); err != nil {
				return nil, err
			}
		}
	}
	return session, nil
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
	agendas, err := u.getLockAgenda(agenda, time.Time{}, time.Time{}, nil)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if agendas == nil {
		return nil, u.error(pkg.ErrPrefInternal, pkg.ErrAgendaNotFound, 0, 0)
	}
	agenda = agendas[0]
	agenda.Status = pkg.AgendaStatusOpenned
	return agenda, nil
}

// tieSession ties session to agendas
func (u *Usecase) tieSession(session *domain.Session) (*domain.Session, error) {
	agendas, err := u.searchLockAgendas(session)
	if err != nil {
		return nil, err
	}
	defer u.unlockAgendas(agendas)
	agenda, err := u.findAgenda(session, agendas)
	if err != nil {
		return nil, err
	}
	over, err := u.getOverlappingSession(agenda)
	if err != nil {
		return nil, err
	}
	u.matchSessionAgenda(session, agenda)
	if err := u.saveSessionAgenda(session, agenda); err != nil {
		return nil, err
	}
	return over, nil
}

// searchLockAgendas searches agendas first for same day and then for a longer period
func (u *Usecase) searchLockAgendas(session *domain.Session) ([]*domain.Agenda, error) {
	ag := domain.Agenda{ClientID: session.ClientID}
	start := time.Date(session.At.Year(), session.At.Month(), session.At.Day(), 0, 0, 0, 0, time.Local)
	end := time.Date(session.At.Year(), session.At.Month(), session.At.Day(), 23, 59, 59, 0, time.Local)
	agendas, err := u.getLockAgenda(&ag, start, end, []string{pkg.AgendaStatusOpenned})
	if err != nil {
		return nil, err
	}
	if agendas == nil {
		start = session.At.Add(-time.Hour * 24 * 60)
		end = session.At.Add(time.Hour * 24 * 60)	
		agendas, err = u.getLockAgenda(&ag, start, end, []string{pkg.AgendaStatusOpenned, pkg.AgendaStatusLocked})
		if err != nil {
			return nil, err
		}
	}
	return agendas, nil
}

// saveSessionAgenda saves the session agenda
func (u *Usecase) saveSessionAgenda(session *domain.Session, agenda *domain.Agenda) error {
	if err := u.Repo.Begin(""); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback("")
	if agenda != nil {
		if err := u.Repo.Save(agenda, ""); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
	}
	if session != nil {
		if err := u.Repo.Save(session, ""); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
	}
	if err := u.Repo.Commit(""); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}

// getLockSession gets a session domain for processing and lock it
func (u *Usecase) getLockSession(id string) (*domain.Session, error) {
	session := &domain.Session{ID: id}
	if err := u.Repo.Begin(""); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback("")
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
	if err := u.Repo.Commit(""); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return session, nil
}

// getLockAgenda gets a agenda based on session params and lock if
func (u *Usecase) getLockAgenda(agenda *domain.Agenda, start time.Time, end time.Time, status []string) ([]*domain.Agenda, error) {
	if err := u.Repo.Begin(""); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback("")
	agendas, err := agenda.LoadRange(u.Repo, start, end, status)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if len(agendas) == 0 {
		return nil, nil
	}
	if err := u.lockAgendas(agendas); err != nil {
		return nil, err
	}
	if err := u.Repo.Commit(""); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return agendas, nil
}

// lockagendas locks slice of agendas
func (u *Usecase) lockAgendas(agendas []*domain.Agenda) error {
	for _, a := range agendas {
		if err := a.Lock(u.Repo, 2); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
	}
	return nil
}

// findAgendas finds agendas linked with session
func (u *Usecase) findAgenda(session *domain.Session, agendas []*domain.Agenda) (*domain.Agenda, error) {
	if session == nil || agendas == nil {
		return nil, nil
	}
	ag := &domain.Agenda{}
	cmd := pkg.Commands{}
	sKeys := []interface{}{session.ClientID, session.At, session.ServiceID}
	weights := []float64{100, 10, 1}
	dist := -1.0
	for _, a := range agendas {
		aKeys := []interface{}{a.ClientID, a.Start, a.ServiceID}
		idx, err := cmd.WeightedDistance(sKeys, aKeys, weights)
		if err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
		if a.Status == pkg.AgendaStatusLocked && a.Start.Format("2006-01-02") != session.At.Format("2006-01-02") {
			continue
		}
		if dist != -1.0 && idx >= dist {
			continue
		}
		ag = a
		dist = idx
	}
	return ag, nil
}

// getOverlappingSession gets overlapping session matched with found agenda
func (u *Usecase) getOverlappingSession(agenda *domain.Agenda) (*domain.Session, error) {
	if agenda == nil || agenda.Status != pkg.AgendaStatusLocked {
		return nil, nil
	}
	session := &domain.Session{AgendaID: agenda.ID}
	if err := u.Repo.Begin(""); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback("")
	i, _, err := u.Repo.Find(session, -1, "")
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	ret := i.(*[]domain.Session)
	if len(*ret) == 0 {
		return nil, u.error(pkg.ErrPrefInternal, pkg.ErrSessionNotFound, 0, 0)
	}
	if len(*ret) > 1 {
		return nil, u.error(pkg.ErrPrefInternal, pkg.ErrSessionMultiples, 0, 0)
	}
	return &(*ret)[0], nil
}

// saveSessionAgenda saves the session agenda
func (u *Usecase) matchSessionAgenda(session *domain.Session, agenda *domain.Agenda) {
	switch {
	case agenda == nil:
		session.Process = pkg.ProcessStatusUnfound
	case session.At.Format("2006-01-02") != agenda.Start.Format("2006-01-02"):
		session.Process = pkg.ProcessStatusUnconfirmed
		session.AgendaID = agenda.ID
		agenda.Status = pkg.AgendaStatusLocked
	case session.ServiceID != agenda.ServiceID:
		session.Process = pkg.ProcessStatusLinked
		session.AgendaID = agenda.ID
		agenda.Status = session.Status
	default:
		session.Process = pkg.ProcessStatusLinked
		session.AgendaID = agenda.ID
		agenda.Status = session.Status
	}
}

// unlock session unlocks session
func (u *Usecase) unlockSession(session *domain.Session) error {
	if err := u.Repo.Begin(""); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback("")
	if err := session.Unlock(u.Repo); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.Repo.Commit(""); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}

// unlock agendas unlocks slice of agendas
func (u *Usecase) unlockAgendas(agendas []*domain.Agenda) error {
	if agendas == nil {
		return nil
	}
	if err := u.Repo.Begin(""); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback("")
	for _, a := range agendas {
		if err := a.Unlock(u.Repo); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
		}
	}
	if err := u.Repo.Commit(""); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}
