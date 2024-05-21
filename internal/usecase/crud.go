package usecase

import (
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Add is a method that add a dto to the repository
func (c *Usecase) Add(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(c.Repo); err != nil {
		return c.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer c.Repo.Rollback()
	domains := in.GetDomain()
	result := []interface{}{}
	count := 1
	for _, domain := range domains {
		if err := domain.Format(c.Repo); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error(), count, len(domains))
		}
		if err := c.Repo.Add(domain); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error(), count, len(domains))
		}
		result = append(result, c.sliceOf(domain))
		count++
	}
	if err := c.Repo.Commit(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	out := in.GetOut()
	c.Out = out.GetDTO(result)
	return nil
}

// Get is a method that gets a dto from the repository
func (c *Usecase) Get(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(c.Repo); err != nil {
		return c.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer c.Repo.Rollback()
	domains := in.GetDomain()
	result := []interface{}{}
	limited := false
	count := 1
	for _, domain := range domains {
		domain, extras, err := in.GetInstructions(domain)
		if err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error(), count, len(domains))
		}
		if err := domain.Format(c.Repo, "filled", "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error(), count, len(domains))
		}
		base, lim, err := c.Repo.Find(domain, pkg.ResultLimit, extras...)
		limited = lim
		if err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error(), count, len(domains))
		}
		if base == nil {
			return c.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound, count, len(domains))
		}
		result = append(result, base)
		count++
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
		return c.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	domains := in.GetDomain()
	result := []interface{}{}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer c.Repo.Rollback()
	count := 1
	for _, source := range domains {
		if err := source.Format(c.Repo, "filled", "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error(), count, len(domains))
		}
		target := source.GetEmpty()
		if f, err := c.Repo.Get(target, source.GetID()); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error(), count, len(domains))
		} else if !f {
			return c.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound, count, len(domains))
		}
		if err := c.merge(source, target); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error(), count, len(domains))
		}
		if err := target.Format(c.Repo, "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error(), count, len(domains))
		}
		if err := c.Repo.Save(target); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error(), count, len(domains))
		}
		result = append(result, c.sliceOf(target))
		count++
	}
	if err := c.Repo.Commit(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	out := in.GetOut()
	c.Out = out.GetDTO(result)
	return nil
}
