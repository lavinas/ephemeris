package usecase

import (
	"reflect"

	"github.com/lavinas/ephemeris/internal/port"
)

// Up is a method that updates a dto in the repository
func (u *Usecase) Up(in port.DTO) (interface{}, string, error) {
	if err := in.Validate(); err != nil {
		err := u.error(ErrPrefBadRequest, err.Error())
		return nil, err.Error(), err
	}
	source := in.GetDomain()
	target := in.GetDomain()
	source.Format()
	if f, err := u.Repo.Get(target, source.GetID()); err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	} else if !f {
		err := u.error(ErrPrefConflict, port.ErrUnfound)
		return nil, err.Error(), err
	}
	if err := u.Merge(source, target); err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	}
	if err := u.Repo.Save(target); err != nil {
		err := u.error(ErrPrefInternal, err.Error())
		return nil, err.Error(), err
	}
	out, strout := in.GetDto(target)
	if out == nil {
		err := u.error(ErrPrefBadRequest, port.ErrUnfound)
		return nil, err.Error(), err
	}
	return out, strout, nil
}

// merge is a method that merges two structs
func (u *Usecase) Merge(source interface{}, target interface{}) error {
	if reflect.TypeOf(source) != reflect.TypeOf(target) {
		return u.error(ErrPrefInternal, port.ErrInvalidTypeOnMerge)
	}
	s := reflect.ValueOf(source).Elem()
	t := reflect.ValueOf(target).Elem()
	for i := 0; i < s.NumField(); i++ {
		if s.Field(i).Interface() != reflect.Zero(s.Field(i).Type()).Interface() {
			t.Field(i).Set(s.Field(i))
		}
	}
	return nil
}
