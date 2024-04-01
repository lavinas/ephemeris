package usecase

import (
	"github.com/lavinas/ephemeris/internal/port"
)

// Get is a method that gets a dto from the repository
func (u *Usecase) Get(in port.DTO) (interface{}, string, error) {
	if err := in.Validate(); err != nil {
		err := u.error(port.ErrPrefBadRequest, err.Error())
		return nil, err.Error(), err
	}
	domain := in.GetDomain()
	if err := domain.Format("filled"); err != nil {
		err := u.error(port.ErrPrefBadRequest, err.Error())
		return nil, err.Error(), err
	}
	found, err := u.Repo.Find(domain)
	if err != nil {
		err := u.error(port.ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	}
	out, strout := in.GetDto(found)
	if out == nil {
		err := u.error(port.ErrPrefBadRequest, port.ErrUnfound)
		return nil, err.Error(), err
	}
	return out, strout, nil
}
