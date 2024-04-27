package usecase

import (
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaMake makes a preview of the agenda based on the client, contract and month
func (u *Usecase) AgendaMake(dtoIn interface{}) error {
	agenda := dtoIn.(*dto.AgendaMake)
	if err := agenda.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	contracts, err := u.getContracts(agenda.ClientID, agenda.ContractID)
	if err != nil {
		return err
	}
	for _, contract := range *contracts {
		if err := contract.Lock(u.Repo); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error())
		}
		// generate agenda
		if err := contract.Unlock(u.Repo); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error())
		}
	}
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