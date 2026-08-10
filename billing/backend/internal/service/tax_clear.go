package service

import (
	"fmt"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// TaxClear is responsible for handling the business logic of receiving emissions.
type TaxClear struct {
	Base
	receiver port.Taxer
}

// NewTaxClear creates a new instance of TaxClear.
func NewTaxClear(repo port.Repository, logger port.Logger, taxer port.Taxer) *TaxClear {
	return &TaxClear{
		Base:     *NewBase(repo, logger),
		receiver: taxer,
	}
}

// Run processes the request to receive emissions and returns the response.
func (s *TaxClear) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(1, "Running TaxClear service")
	receiveDTO, ok := inDTO.(*dto.TaxClearRequest)
	if !ok {
		s.logger.IPrintf(1, "Invalid input type")
		return dto.NewTaxClearResponse(400, "error", "Invalid input type")
	}
	// Validate the input DTO
	if err := receiveDTO.Validate(s.repo); err != nil {
		s.logger.IPrintf(1, "Validation error: %v", err)
		return dto.NewTaxClearResponse(400, "error", err.Error())
	}
	// Call the receiver to process the emission receive
	fileItems, err := s.receiver.ClearEmission(receiveDTO.Base64Source)
	if err != nil {
		s.logger.IPrintf(1, "Error receiving emission: %v", err)
		return dto.NewTaxClearResponse(400, "error", err.Error())
	}
	// Get the domain emission from the DTO
	emission := receiveDTO.GetDomain()
	// Merge the received emission items with the existing emission items
	if err := s.mergeEmissionItems(emission, fileItems); err != nil {
		s.logger.IPrintf(1, "Error merging emission items: %v", err)
		return dto.NewTaxClearResponse(400, "error", err.Error())
	}
	// Save the updated emission to the repository
	if err := s.saveEmission(emission); err != nil {
		s.logger.IPrintf(1, "Error saving emission: %v", err)
		return dto.NewTaxClearResponse(500, "error", "Failed to save emission")
	}
	// Process the emission receive logic here
	s.logger.IPrintf(1, "Emission received successfully: ID %d, Quantity %d, Amount %.2f",
		emission.ID, emission.Quantity, emission.Amount)
	// Return a successful response
	return dto.NewTaxClearResponse(200, "success", "Emission received successfully")
}

// mergeEmissionItems merges the received emission items with the existing emission items in the domain model.
func (s *TaxClear) mergeEmissionItems(emission *domain.Emission,
	fileItems map[int64]*domain.EmissionItem) error {
	if emission.NFEDatetime != nil {
		return fmt.Errorf("emission has already been processed with NFE datetime: %v", *emission.NFEDatetime)
	}
	if len(fileItems) != len(emission.EmissionItems) {
		return fmt.Errorf("mismatch in number of emission items: expected %d, got %d",
			len(emission.EmissionItems), len(fileItems))
	}
	for i, item := range emission.EmissionItems {
		fileItem, exists := fileItems[item.RPSNumber]
		if !exists {
			return fmt.Errorf("emission item with RPS number %d not found in received items", item.RPSNumber)
		}
		// Update the emission item with the received data
		emission.EmissionItems[i].NFENumber = fileItem.NFENumber
		emission.EmissionItems[i].NFEDatetime = fileItem.NFEDatetime
		emission.EmissionItems[i].NFEVerification = fileItem.NFEVerification
		emission.EmissionItems[i].NFEAmount = fileItem.NFEAmount
	}
	emission.UpdatedAt = time.Now()
	emission.NFEStart = fileItems[emission.EmissionItems[0].RPSNumber].NFENumber
	emission.NFEEnd = fileItems[emission.EmissionItems[len(emission.EmissionItems)-1].RPSNumber].NFENumber
	emission.NFEDatetime = fileItems[emission.EmissionItems[len(emission.EmissionItems)-1].RPSNumber].NFEDatetime
	s.logger.IPrintf(2, "Updated emission NFE range: Start %d, End %d",
		*emission.NFEStart, *emission.NFEEnd)
	s.logger.IPrintf(2, "Successfully merged %d emission items for emission ID %d",
		len(emission.EmissionItems), emission.ID)
	return nil
}

// saveEmission saves the updated emission to the repository.
func (s *TaxClear) saveEmission(emission *domain.Emission) error {
	if err := s.repo.BeginTransaction(); err != nil {
		s.logger.IPrintf(1, "Error starting transaction: %v", err)
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer s.repo.RollbackTransaction()
	for _, item := range emission.EmissionItems {
		if err := s.repo.Save(item); err != nil {
			s.logger.IPrintf(1, "Error saving emission item: %v", err)
			return fmt.Errorf("failed to save emission item: %v", err)
		}
	}
	if err := s.repo.Save(emission); err != nil {
		s.logger.IPrintf(1, "Error saving emission: %v", err)
		return fmt.Errorf("failed to save emission: %v", err)
	}
	if err := s.repo.CommitTransaction(); err != nil {
		s.logger.IPrintf(1, "Error committing transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	s.logger.IPrintf(1, "Successfully saved emission with ID %d and its items", emission.ID)
	return nil
}
