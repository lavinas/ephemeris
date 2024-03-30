package usecase

import (
	"github.com/lavinas/ephemeris/internal/port"
)

// Up is a method that updates a dto in the repository
func (u *Usecase) Up(in port.DTO) (interface{}, string, error) {
	if err := in.Validate(); err != nil {
		err := u.error(ErrPrefBadRequest, err.Error())
		return nil, err.Error(), err
	}
	domain := in.GetDomain()
	domain.Format()

	if f, err := u.Repo.Get(domain, domain.GetID()); err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	} else if !f {
		err := u.error(ErrPrefConflict, port.ErrUnfound)
		return nil, err.Error(), err
	}

	return nil, "", nil
}
