package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaMake makes a preview of the agenda based on the client, contract and month
func (u *Usecase) AgendaMake(dtoIn dto.AgendaMake) error {
	if err := dtoIn.Validate(); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	contracts, err := u.GetContracts(dtoIn.ClientID, dtoIn.ContractID)
	if err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	for _, contract := range *contracts {
		fmt.Println(contract)
	}

	// month := dtoIn.GetMonth()
	// lock agenda for client and contract in the month
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	defer u.Repo.Rollback()
	// delete all the agenda items for the client and contract in the month
	// insert the new agenda items
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	// Liberate the lock
	// generate output
	return nil
}

// deleteAgenda deletes Agenda based on client, contract and month
func (u *Usecase) DeleteAgenda(clientID, contractID int, month time.Time) error {
	return nil
}


// GetClientContracts is a method that returns all contracts of a client
func (u *Usecase) GetContracts(clientID, contractID string) (*[]domain.Contract, error) {
	if contractID == "" && clientID == "" {
		return nil, errors.New(pkg.ErrClientContractEmpty)
	}
	contract := &domain.Contract{}
	if clientID != "" {
		contract.ClientID = clientID
	}
	if contractID != "" {
		contract.ID = contractID
	}
	ret, _, err := u.Repo.Find(contract, -1)
	if err != nil {
		return nil, err
	}
	return ret.(*[]domain.Contract), nil
}