package usecase

import (
	"fmt"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaMatch makes a match between the agenda and the sessions
func (u *Usecase) AgendaMatch(dtoIn interface{}) error {
	in := dtoIn.(*dto.AgendaMatch)
	if err := in.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	sessions, err := u.getSessionsMatch(in)
	if err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	agendas, err := u.getAgendaMatch(in)
	if err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.matchAgendasSessions(sessions, agendas); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}

// getSessionsMatch returns the sessions based on the client and month
func (u *Usecase) getSessionsMatch(dto *dto.AgendaMatch) (map[string]*domain.Session, error) {
	mapSessions := map[string]*domain.Session{}
	s, inst, err := dto.GetMatchDomain("session")
	if err != nil {
		return nil, err
	}
	sessions, _, err := u.Repo.Find(s, 0, inst...)
	if err != nil {
		return nil, err
	}
	for _, session := range *sessions.(*[]domain.Session) {
		key := fmt.Sprintf("%s-%s-%s", session.ClientID, session.ServiceID, session.At.Format("2006-01-02"))
		mapSessions[key] = &session
	}
	return mapSessions, nil
}

// getAgendaMatch returns the agenda based on the client and month
func (u *Usecase) getAgendaMatch(dto *dto.AgendaMatch) (map[string]*domain.Agenda, error) {
	mapSessions := map[string]*domain.Agenda{}
	s, inst, err := dto.GetMatchDomain("agenda")
	if err != nil {
		return nil, err
	}
	agendas, _, err := u.Repo.Find(s, 0, inst...)
	if err != nil {
		return nil, err
	}
	for _, agenda := range *agendas.(*[]domain.Agenda) {
		key := fmt.Sprintf("%s-%s-%s", agenda.ClientID, agenda.ServiceID, agenda.Start.Format("2006-01-02"))
		mapSessions[key] = &agenda
	}
	return mapSessions, nil
}

// matchAgendaSession matches the agenda with the session
func (u *Usecase) matchAgendasSessions(sessions map[string]*domain.Session, agendas map[string]*domain.Agenda) error {
	for key := range sessions {
		fmt.Println(1, key)
	}
	for key := range agendas {
		fmt.Println(2, key)
	}
	return nil
}
