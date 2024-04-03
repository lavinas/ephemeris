package usecase

import (
	"fmt"
	"github.com/lavinas/ephemeris/internal/port"
)

// Add is a method that add a dto to the repository
func (u *Usecase) Add(in port.DTOIn) ([]port.DTOOut, string, error) {
	if err := in.Validate(); err != nil {
		err := u.error(port.ErrPrefBadRequest, err.Error())
		return nil, err.Error(), err
	}
	domain := in.GetDomain()
	if err := domain.Format(); err != nil {
		err := u.error(port.ErrPrefBadRequest, err.Error())
		return nil, err.Error(), err
	}
	if f, err := u.Repo.Get(domain, domain.GetID()); err != nil {
		err := u.error(port.ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	} else if f {
		err := u.error(port.ErrPrefConflict, fmt.Sprintf(port.ErrAlreadyExists, domain.GetID()))
		return nil, err.Error(), err
	}
	if err := u.Repo.Add(domain); err != nil {
		err := u.error(port.ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	}
	
	out, strout := in.GetOut(domain)
	return out, strout, nil
}
