package usecase

import (
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
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

// GetClientContracts is a method that returns all contracts of a client
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
	ret := []port.DTOOut{}
	if err := u.Repo.Begin(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	defer u.Repo.Rollback()
	dtoOut := dtoIn.GetOut()
	starts, ends, err := u.getDates(contract, month)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	for i := 0; i < len(starts); i++ {
		agenda := dtoIn.GetDomain()[0].(*domain.Agenda)
		u.setDates(agenda, contract.ID, starts[i], ends[i])
		if err := agenda.Format(u.Repo); err != nil {
			return nil, u.error(pkg.ErrPrefBadRequest, err.Error())
		}
		if err := u.Repo.Add(agenda); err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error())
		}
		ret = append(ret, dtoOut.GetDTO(agenda)...)
	}
	if err := u.Repo.Commit(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	return ret, nil
}

// getDates returns the dates of the contract based on the month
func (u *Usecase) getDates(contract *domain.Contract, month time.Time) ([]time.Time, []time.Time, error) {
	beginMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	endMonth := beginMonth.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	starts := []time.Time{}
	ends := []time.Time{}
	recur, minutes, err := u.getPackageParams(contract.PackageID)
	if err != nil {
		return nil, nil, err
	}
	for start := &contract.Start; start != nil && !start.After(endMonth); start = recur.Next(*start) {
		if !start.Before(beginMonth) && !start.After(endMonth) {
			starts = append(starts, *start)
			ends = append(ends, start.Add(time.Minute*time.Duration(minutes)))
		}
		if recur.Next(*start) == nil {
			break
		}
	}
	return starts, ends, nil
}

// getPackageParams returns the recurrence struct and serviice minutes of the package
func (u *Usecase) getPackageParams(packId string) (*domain.Recurrence, int, error) {
	pack := domain.Package{ID: packId}
	var err error
	recur, err := pack.GetRecurrence(u.Repo)
	if err != nil {
		return nil, 0, u.error(pkg.ErrPrefInternal, err.Error())
	}
	service, err := pack.GetService(u.Repo)
	if err != nil {
		return nil, 0, u.error(pkg.ErrPrefInternal, err.Error())
	}
	var minutes int64 = 0
	if service.Minutes != nil {
		minutes = *service.Minutes
	}
	return recur, int(minutes), nil
}

// setDates sets the dates of the agenda and id based on dates and contract and month
func (u *Usecase) setDates(agenda *domain.Agenda, contractID string, start time.Time, end time.Time) {
	agenda.ID = fmt.Sprintf("%s-%s", contractID, start.Format(pkg.DateFormat))
	agenda.Start = start
	agenda.End = end
}