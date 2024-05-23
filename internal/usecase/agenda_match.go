package usecase

import (
	"time"
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
	month, _ := time.Parse(pkg.MonthFormat, in.Month)
	sessions, err := u.getWaitingSessions(in.ClientID, in.ContractID, month)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		fmt.Println(session)
	}
	
	return nil
}


// getSessions returns the sessions for the given month
func (u *Usecase) getWaitingSessions(client, contract string, month time.Time) ([]*domain.Session, error) {
	return nil, nil
}