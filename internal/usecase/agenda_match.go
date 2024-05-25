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
	agenda := dto.GetDomain()[0].(*domain.Agenda)
	instructions := dto.GetInstructions()
	sessions, _, err := u.Repo.Find(agenda, 0, instructions)
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
	mapAgendas := map[string]*domain.Agenda{}
	agenda := dto.GetDomain()[0].(*domain.Agenda)
	instructions := dto.GetInstructions()
	agendas, _, err := u.Repo.Find(agenda, 0, instructions)
	if err != nil {
		return nil, err
	}
	for _, agenda := range *agendas.(*[]domain.Agenda) {
		key := fmt.Sprintf("%s-%s", agenda.ClientID, agenda.Kind)
		mapAgendas[key] = &agenda
	}
	return mapAgendas, nil
}

// matchAgendaSession matches the agenda with the session
func (u *Usecase) matchAgendasSessions(sessions map[string]*domain.Session, agendas map[string]*domain.Agenda) error {
	fmt.Println(sessions)
	fmt.Println(agendas)
	return nil
}
