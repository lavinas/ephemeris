package usecase

import (
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
		if contract.IsLocked() {
			ret = append(ret, &dto.AgendaMakeOut{
				ID:         "",
				ClientID:   contract.ClientID,
				ContractID: contract.ID,
				Start:      pkg.Locked,
				End:        pkg.Locked,
			})
			continue
		}
		if err := contract.Lock(u.Repo); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error())
		}
		u.generateAgenda(&contract, month)
		if err := contract.Unlock(u.Repo); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error())
		}
	}
	u.Out = ret
	return nil
}

// deleteAgenda deletes Agenda based on client, contract and month
func (u *Usecase) DeleteAgenda(clientID, contractID int, month time.Time) error {
	return nil
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

// generateAgenda generates the agenda based on the contract
func (u *Usecase) generateAgenda(contract *domain.Contract, month time.Time) *[]domain.Agenda {
	return nil
}