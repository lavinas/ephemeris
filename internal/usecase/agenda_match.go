package usecase

import (

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaMatch makes a match between the agenda and the sessions
func (u *Usecase) AgendaMatch(dtoIn interface{}) error {
	in := dtoIn.(*dto.AgendaMatch)
	if err := in.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	return nil
}

