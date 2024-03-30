package usecase

import (
	"fmt"
	"github.com/lavinas/ephemeris/internal/port"
)

// Add is a method that add a dto to the repository
func (u *Usecase) Add(in port.DTO) (interface{}, string, error) {
	if err := in.Validate(); err != nil {
		err := u.error(ErrPrefBadRequest, err.Error())
		return nil, err.Error(), err
	}
	domain := in.GetDomain()
	domain.Format()
	if f, err := u.Repo.Get(domain, domain.GetID()); err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	} else if f {
		err := u.error(ErrPrefConflict, fmt.Sprintf(port.ErrAlreadyExists, domain.GetID()))
		return nil, err.Error(), err
	}
	if err := u.Repo.Add(domain); err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	}
	out, strout := in.GetDto(domain)
	return out, strout, nil
}
