package usecase

import (
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

const (
	idDateFormat = "2006-01-02-15"
	idFormat     = "%s-%s"
)

// AgendaMake makes a preview of the agenda based on the client, contract and month
func (u *Usecase) AgendaMake(dtoIn interface{}) error {
	dtoAgenda := dtoIn.(*dto.AgendaMake)
	if err := dtoAgenda.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	month, _ := time.Parse(pkg.MonthFormat, dtoAgenda.Month)
	contracts, err := u.getContracts(dtoAgenda.ClientID, dtoAgenda.ContractID)
	if err != nil {
		return err
	}
	ret := []port.DTOOut{}
	for _, contract := range *contracts {
		out, err := u.AgendaContractMake(dtoAgenda, contract, month)
		if err != nil {
			return err
		}
		ret = append(ret, out...)
	}
	u.Out = ret
	return nil
}

// AgendaContractMake makes a preview of the agenda based on the client, contract and month
func (u *Usecase) AgendaContractMake(dtoIn port.DTOIn, contract domain.Contract, month time.Time) ([]port.DTOOut, error) {
	if contract.IsLocked() {
		ret := dto.AgendaMakeOut{ID: "", ClientID: contract.ClientID, ContractID: contract.ID,
			Start: pkg.Locked, End: pkg.Locked, Kind: pkg.Locked, Status: pkg.Locked}
		return []port.DTOOut{&ret}, nil
	}
	if err := contract.Lock(u.Repo); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	defer contract.Unlock(u.Repo)
	if err := u.DeleteAgenda(&contract, month); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	ret, err := u.GenerateAgenda(dtoIn, &contract, month)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	return ret, nil
}

// deleteAgenda deletes Agenda based on client, contract, month and status
func (u *Usecase) DeleteAgenda(contract *domain.Contract, month time.Time) error {
	firstday := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	lastday := firstday.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	defer u.Repo.Rollback()
	agenda := &domain.Agenda{ContractID: contract.ID}
	if err := u.Repo.Delete(agenda, "start >= ? AND start <= ?", firstday, lastday); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	return nil
}

// generateAgenda generates the agenda based on the contract
func (u *Usecase) GenerateAgenda(dtoIn port.DTOIn, contract *domain.Contract, month time.Time) ([]port.DTOOut, error) {
	if err := u.Repo.Begin(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	defer u.Repo.Rollback()
	starts, ends, err := u.getDates(contract, month)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	agendas, err := u.getAgenda(dtoIn, contract, starts, ends)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	if err := u.Repo.Commit(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	return agendas, nil
}

// GetContracts is a method that returns all contracts of a client
func (u *Usecase) getContracts(clientID, contractID string) (*[]domain.Contract, error) {
	if contractID == "" && clientID == "" {
		return nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrClientContractEmpty)
	}
	contract := &domain.Contract{}
	if clientID != "" {
		contract.ClientID = clientID
	}
	if contractID != "" {
		contract.ID = contractID
	}
	ret, _, err := u.Repo.Find(contract, 100)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	if ret == nil {
		return nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound)
	}
	return ret.(*[]domain.Contract), nil
}

// getAgenda generates the agenda based on the contract
func (u *Usecase) getAgenda(dtoIn port.DTOIn, contract *domain.Contract, starts []time.Time, ends []time.Time) ([]port.DTOOut, error) {
	ret := []port.DTOOut{}
	agenda := dtoIn.GetDomain()[0].(*domain.Agenda)
	dtoOut := dtoIn.GetOut()
	for i := 0; i < len(starts); i++ {
		agenda.ContractID = contract.ID
		agenda.ClientID = contract.ClientID
		u.setDates(agenda, contract.ClientID, starts[i], ends[i])
		if err := agenda.Format(u.Repo); err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error())
		}
		if err := u.Repo.Add(agenda); err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error())
		}
		ret = append(ret, dtoOut.GetDTO(agenda)...)
	}
	return ret, nil
}

// getDates returns the dates of the contract based on the month
func (u *Usecase) getDates(contract *domain.Contract, month time.Time) ([]time.Time, []time.Time, error) {
	starts, ends, err := u.mountDates(contract, month)
	if err != nil {
		return nil, nil, err
	}
	starts, ends, err = u.delBound(contract, month, starts, ends)
	if err != nil {
		return nil, nil, err
	}
	return starts, ends, nil
}

// mountDates returns the dates of the contract based on the month
func (u *Usecase) mountDates(contract *domain.Contract, month time.Time) ([]time.Time, []time.Time, error) {
	beginMonth, endMonth := u.getMonthBound(contract, month)
	recur, services, err := u.getPackageParams(contract.PackageID)
	if err != nil {
		return nil, nil, err
	}
	starts := []time.Time{}
	ends := []time.Time{}
	count := 0
	appended := 0
	for start := &contract.Start; start != nil && !start.After(endMonth); start = recur.Next(*start) {
		m := services[count%len(services)].Minutes
		count++
		var minutes int64 = 0
		if m != nil {
			minutes = *m
		}
		if !start.Before(beginMonth) && !start.After(endMonth) {
			starts = append(starts, *start)
			ends = append(ends, start.Add(time.Minute*time.Duration(minutes)))
			appended++
		}
		if recur.Limits != nil && appended >= int(*recur.Limits) {
			break
		}
	}
	return starts, ends, nil
}

// getMonthBound returns the bound of the contract based on the month
func (u *Usecase) getMonthBound(contract *domain.Contract, month time.Time) (time.Time, time.Time) {
	beginMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	endMonth := beginMonth.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	if contract.End != nil && contract.End.Before(endMonth) {
		endMonth = *contract.End
		endMonth = endMonth.AddDate(0, 0, 1).Add(time.Nanosecond * -1)
	}
	return beginMonth, endMonth
}

// delBound deletes the bound of the contract
func (u *Usecase) delBound(contract *domain.Contract, month time.Time, starts, ends []time.Time) ([]time.Time, []time.Time, error) {
	bond, err := contract.GetBond(u.Repo)
	if err != nil {
		return nil, nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	if bond == nil {
		return starts, ends, nil
	}
	delStarts, _, err := u.getDates(bond, month)
	if err != nil {
		return nil, nil, err
	}
	st, pos := u.minus(starts, delStarts)
	return st, u.keep(ends, pos), nil
}

// minus returns the subtracted slice minus the subtractor slice
func (u *Usecase) minus(subtracted []time.Time, subtractor []time.Time) ([]time.Time, []int) {
	times := []time.Time{}
	pos := []int{}
	maps := make(map[time.Time]bool)
	for _, s := range subtractor {
		maps[s] = true
	}
	count := 0
	for _, s := range subtracted {
		if _, ok := maps[s]; !ok {
			times = append(times, s)
			pos = append(pos, count)
		}
		count++
	}
	return times, pos
}

// keep returns the slice with the positions
func (u *Usecase) keep(times []time.Time, pos []int) []time.Time {
	ret := []time.Time{}
	for _, p := range pos {
		ret = append(ret, times[p])
	}
	return ret
}

// getPackageParams returns the recurrence struct and serviice minutes of the package
func (u *Usecase) getPackageParams(packId string) (*domain.Recurrence, []*domain.Service, error) {
	pack := domain.Package{ID: packId}
	var err error
	recur, err := pack.GetRecurrence(u.Repo)
	if err != nil {
		return nil, []*domain.Service{}, u.error(pkg.ErrPrefInternal, err.Error())
	}
	services, err := pack.GetService(u.Repo)
	if err != nil {
		return nil, []*domain.Service{}, u.error(pkg.ErrPrefInternal, err.Error())
	}
	return recur, services, nil
}

// setDates sets the dates of the agenda and id based on dates and contract and month
func (u *Usecase) setDates(agenda *domain.Agenda, clientID string, start time.Time, end time.Time) {
	agenda.ID = fmt.Sprintf(idFormat, start.Format(idDateFormat), clientID)
	agenda.Start = start
	agenda.End = end
}
