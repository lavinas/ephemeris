package usecase

import (
	"fmt"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

const (
	ErrWrongAddClientDTO   = "internal error: wrong AddClient dto"
	ErrWrongGetClientDTO   = "internal error: wrong GetClient dto"
	ErrClientAlreadyExists = "conflict: client already exists with id %s"
	ErrGetNotFound         = "not found: registers unfound with the informed params"
)

// Add is a method that add a client to the repository
func (u *Usecase) Add(in port.DTO) (interface{}, string, error) {
	domain := in.GetDomain()
	if err := domain.Validate(); err != nil {
		err := u.error(ErrPrefBadRequest, err.Error())
		return nil, err.Error(), err
	}
	domain.Format()
	if f, err := u.Repo.Get(domain, domain.GetID()); err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	} else if f {
		err := u.error(ErrPrefConflict, fmt.Sprintf(ErrClientAlreadyExists, domain.GetID()))
		return nil, err.Error(), err
	}
	if err := u.Repo.Add(domain); err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	}
	return nil, "ok: client added", nil
}

// Get is a method that gets a client from the repository
func (u *Usecase) Get(in port.DTO) (interface{}, string, error) {
	client := in.GetDomain()
	client.Format()
	if in.IsEmpty() {
		err := u.error(ErrPrefBadRequest, "no params is informed")
		return nil, err.Error(), err
	}
	clients, err := u.Repo.Find(client)
	if err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	}
	out := in.GetDto(clients)
	if out == nil {
		err := u.error(ErrGetNotFound, ErrGetNotFound)
		return nil, err.Error(), err
	}
	comm := pkg.NewCommands()
	x := comm.Marshal(out, "nokeys")
	return out, x, nil
}
