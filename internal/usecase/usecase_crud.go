package usecase

import (
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Add is a method that add a dto to the repository
func (c *Usecase) Add(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(c.Repo); err != nil {
		return c.error(pkg.ErrPrefBadRequest, err.Error())
	}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	defer c.Repo.Rollback()
	domains := in.GetDomain()
	result := []interface{}{}
	for _, domain := range domains {
		if err := domain.Format(c.Repo); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error())
		}
		if err := c.Repo.Add(domain); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		result = append(result, c.sliceOf(domain))
	}
	if err := c.Repo.Commit(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	out := in.GetOut()
	c.Out = out.GetDTO(result)
	return nil
}

// Get is a method that gets a dto from the repository
func (c *Usecase) Get(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(c.Repo); err != nil {
		return c.error(pkg.ErrPrefBadRequest, err.Error())
	}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	defer c.Repo.Rollback()
	domains := in.GetDomain()
	result := []interface{}{}
	limited := false
	for _, domain := range domains {
		if err := domain.Format(c.Repo, "filled", "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error())
		}
		base, lim, err := c.Repo.Find(domain, pkg.ResultLimit)
		limited = lim
		if err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		if base == nil {
			return c.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound)
		}
		result = append(result, base)
	}
	out := in.GetOut()
	c.Out = out.GetDTO(result)
	c.Limited = limited
	return nil
}

// Up is a method that updates a dto in the repository
func (c *Usecase) Up(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(c.Repo); err != nil {
		return c.error(pkg.ErrPrefBadRequest, err.Error())
	}
	domains := in.GetDomain()
	result := []interface{}{}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	defer c.Repo.Rollback()
	for _, source := range domains {
		if err := source.Format(c.Repo, "filled", "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error())
		}
		target := source.GetEmpty()
		if f, err := c.Repo.Get(target, source.GetID()); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		} else if !f {
			return c.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound)
		}
		if err := c.merge(source, target); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		if err := target.Format(c.Repo, "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		if err := c.Repo.Save(target); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		result = append(result, c.sliceOf(target))
	}
	if err := c.Repo.Commit(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	out := in.GetOut()
	c.Out = out.GetDTO(result)
	return nil
}
