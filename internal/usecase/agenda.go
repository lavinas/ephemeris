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
		out, err := u.AgendaContractMake(contract, month)
		if err != nil {
			return err
		}
		ret = append(ret, out...)
	}
	u.Out = ret
	return nil
}

// AgendaContractMake makes a preview of the agenda based on the client, contract and month
func (u *Usecase) AgendaContractMake(contract domain.Contract, month time.Time) ([]port.DTOOut, error) {
	if contract.IsLocked() {
		ret := dto.AgendaMakeOut{ID: "", ClientID: contract.ClientID, ContractID: contract.ID,
			Start: pkg.Locked, End: pkg.Locked}
		return []port.DTOOut{&ret}, nil
	}
	if err := contract.Lock(u.Repo); err != nil {
		return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	defer contract.Unlock(u.Repo)
	if err := u.DeleteAgenda(&contract, month); err != nil {
	   return nil, u.error(pkg.ErrPrefInternal, err.Error())
	}
	ret, err := u.GenerateAgenda(&contract, month)
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
	for i := firstday; i.Before(lastday); i = i.AddDate(0, 0, 1) {
		agenda := &domain.Agenda{ContractID: contract.ID, 
			                     Start: i, 
								 Status: pkg.AgendaStatusSlated, 
						         Kind: pkg.AgendaKindSlated,
								}
		if err := u.Repo.Delete(agenda); err != nil {
			u.Repo.Rollback()
			return u.error(pkg.ErrPrefInternal, err.Error())
		}
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	return nil
}

// generateAgenda generates the agenda based on the contract
func (u *Usecase) GenerateAgenda(contract *domain.Contract, month time.Time) ([]port.DTOOut, error) {
	firstday := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	lastday := firstday.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	ret := []port.DTOOut{}
	// if err := u.Repo.Begin(); err != nil {
	//	return nil, u.error(pkg.ErrPrefInternal, err.Error())
	// }
	for i := firstday; i.Before(lastday); i = i.AddDate(0, 0, 1) {
		agenda := &domain.Agenda{
			ID:         fmt.Sprintf("%s-%s", contract.ID, i.Format(pkg.DateFormat)),
			ContractID: contract.ID,
			Start:      i,
			End:        i.AddDate(0, 0, 1).Add(time.Nanosecond * -1),
			Kind:       pkg.AgendaKindSlated,
			Status:     pkg.AgendaStatusSlated,
		}
		if err := agenda.Format(u.Repo); err != nil {
			return nil, u.error(pkg.ErrPrefBadRequest, err.Error())
		}
		if err := u.Repo.Begin(); err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error())
		}
		if err := u.Repo.Add(agenda); err != nil {
			u.Repo.Rollback()
			return nil, u.error(pkg.ErrPrefInternal, err.Error())
		}
		if err := u.Repo.Commit(); err != nil {
			return nil, u.error(pkg.ErrPrefInternal, err.Error())
		}
		ret = append(ret, &dto.AgendaMakeOut{ID: agenda.ID,
			ClientID:   contract.ClientID,
			ContractID: contract.ID,
			Start:      agenda.Start.Format(pkg.DateFormat),
			End:        agenda.End.Format(pkg.DateFormat)})
		if i == firstday {
			break
		}
	}
	// if err := u.Repo.Commit(); err != nil {
	//	return nil, u.error(pkg.ErrPrefInternal, err.Error())
	// }
	return ret, nil
}
