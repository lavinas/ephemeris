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

type agendaItem struct {
	start     time.Time
	end       time.Time
	serviceId string
	Price     *float64
}

// AgendaMake makes a preview of the agenda based on the client, contract and month
func (u *Usecase) AgendaMake(dtoIn interface{}) error {
	dtoAgenda := dtoIn.(*dto.AgendaMake)
	if err := dtoAgenda.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error(), 0, 0)
	}
	month, _ := time.Parse(pkg.MonthFormat, dtoAgenda.Month)
	contracts, err := u.getContracts(dtoAgenda)
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
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer contract.Unlock(u.Repo)
	if err := u.DeleteAgenda(&contract, month); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	dtosOut, err := u.GenerateAgenda(dtoIn, &contract, month)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return dtosOut, nil
}

// DeleteAgenda deletes Agenda based on client, contract, month and status
func (u *Usecase) DeleteAgenda(contract *domain.Contract, month time.Time) error {
	firstday := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	lastday := firstday.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	agenda := &domain.Agenda{ContractID: &contract.ID}
	p1 := fmt.Sprintf("start >= '%s'", firstday.Format("2006-01-02 15:04:05"))
	p2 := fmt.Sprintf("start <= '%s'", lastday.Format("2006-01-02 15:04:05"))
	if err := u.Repo.Delete(agenda, p1, p2); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return nil
}

// generateAgenda generates the agenda based on the contract
func (u *Usecase) GenerateAgenda(dtoIn port.DTOIn, contract *domain.Contract, month time.Time) ([]port.DTOOut, error) {
	if err := u.Repo.Begin(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	defer u.Repo.Rollback()
	items, err := u.getItems(contract, month)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	dtosOut, err := u.saveAgenda(dtoIn, contract, items)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if err := u.Repo.Commit(); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return dtosOut, nil
}

// GetContracts is a method that returns all contracts of a client
func (u *Usecase) getContracts(dtoAgenda *dto.AgendaMake) (*[]domain.Contract, error) {
	contract := dtoAgenda.GetDomain()[0].(*domain.Contract)
	_, inst, err := dtoAgenda.GetInstructions(contract)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	ret, _, err := u.Repo.Find(contract, 0, inst...)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if ret == nil {
		return nil, u.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound, 0, 0)
	}
	return ret.(*[]domain.Contract), nil
}

// saveAgenda saves the agenda based on the contract, items and dto
func (u *Usecase) saveAgenda(dtoIn port.DTOIn, contract *domain.Contract, items []*agendaItem) ([]port.DTOOut, error) {
	ret := []port.DTOOut{}
	agenda := domain.Agenda{}
	agenda.Date = time.Now()
	agenda.Kind = pkg.DefaultAgendaKind
	agenda.Status = pkg.DefaultAgendaStatus
	dtoOut := dtoIn.GetOut()
	count := 1
	for i := 0; i < len(items); i++ {
		if err := u.setAgenda(&agenda, contract, items[i]); err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error(), count, len(items))
		}
		if err := u.Repo.Add(agenda); err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error(), count, len(items))
		}
		ret = append(ret, dtoOut.GetDTO(&agenda)...)
		count++
	}
	return ret, nil
}

// getItems returns the items of the agenda based on contract and the month
func (u *Usecase) getItems(contract *domain.Contract, month time.Time) ([]*agendaItem, error) {
	items, err := u.mountItems(contract, month)
	if err != nil {
		return nil, err
	}
	items, err = u.delBound(contract, month, items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// mounItems mounts the agenda items based on the contract and month
func (u *Usecase) mountItems(contract *domain.Contract, month time.Time) ([]*agendaItem, error) {
	beginMonth, endMonth := u.getBound(contract, month)
	recur, services, prices, err := u.getPackageParams(contract.PackageID)
	if err != nil {
		return nil, err
	}
	items := []*agendaItem{}
	count := 0
	appended := 0
	for start := &contract.Start; start != nil && !start.After(endMonth); start = recur.Next(*start) {
		minutes, serviceId, price := u.getServicePrice(services, prices, count)
		if !start.Before(beginMonth) && !start.After(endMonth) {
			end := start.Add(time.Minute * time.Duration(minutes))
			items = append(items, &agendaItem{start: *start, end: end, serviceId: serviceId, Price: price})
			appended++
		}
		if recur.Limits != nil && appended >= int(*recur.Limits) {
			break
		}
		count++
	}
	return items, nil
}

// getBound returns the bound of the contract based on the month
func (u *Usecase) getBound(contract *domain.Contract, month time.Time) (time.Time, time.Time) {
	beginMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	endMonth := beginMonth.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	if contract.End != nil && contract.End.Before(endMonth) {
		endMonth = *contract.End
		endMonth = endMonth.AddDate(0, 0, 1).Add(time.Nanosecond * -1)
	}
	return beginMonth, endMonth
}

// getPackageParams returns the recurrence struct and serviice minutes of the package
func (u *Usecase) getPackageParams(packId string) (*domain.Recurrence, []*domain.Service, []*float64, error) {
	pack := domain.Package{ID: packId}
	if ok, err := pack.Load(u.Repo); err != nil {
		return nil, nil, nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	} else if !ok {
		return nil, nil, nil, u.error(pkg.ErrPrefInternal, pkg.ErrPackageNotFound, 0, 0)
	}
	var err error
	recur, err := pack.GetRecurrence(u.Repo)
	if err != nil {
		return nil, nil, nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	services, prices, err := pack.GetServices(u.Repo)
	if err != nil {
		return nil, nil, nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	return recur, services, prices, nil
}

// getServicePrice returns the service price based on the count
func (u *Usecase) getServicePrice(services []*domain.Service, prices []*float64, count int) (int, string, *float64) {
	idx := count % len(services)
	m := services[idx].Minutes
	var minutes int64 = 0
	if m != nil {
		minutes = *m
	}
	return int(minutes), services[idx].ID, prices[idx]
}

// delBound deletes the bound of the contract
func (u *Usecase) delBound(contract *domain.Contract, month time.Time, items []*agendaItem) ([]*agendaItem, error) {
	bond, err := contract.GetBond(u.Repo)
	if err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error(), 0, 0)
	}
	if bond == nil {
		return items, nil
	}
	delItems, err := u.getItems(bond, month)
	if err != nil {
		return nil, err
	}
	items = u.minus(items, delItems)
	return items, nil
}

// minus returns the subtracted slice minus the subtractor slice
func (u *Usecase) minus(subtracted []*agendaItem, subtractor []*agendaItem) []*agendaItem {
	sub := []*agendaItem{}
	maps := make(map[time.Time]bool)
	for _, s := range subtractor {
		maps[s.start] = true
	}
	count := 0
	for _, s := range subtracted {
		if _, ok := maps[s.start]; !ok {
			sub = append(sub, s)
		}
		count++
	}
	return sub
}

// setAgenda sets the agenda based on the contract
func (u *Usecase) setAgenda(agenda *domain.Agenda, contract *domain.Contract, item *agendaItem) error {
	agenda.ContractID = &contract.ID
	agenda.ClientID = contract.ClientID
	agenda.Start = item.start
	agenda.End = item.end
	agenda.ServiceID = item.serviceId
	agenda.Price = item.Price
	agenda.ID = fmt.Sprintf(idFormat, item.start.Format(idDateFormat), contract.ClientID)
	if err := agenda.Format(u.Repo); err != nil {
		return err
	}
	return nil
}
