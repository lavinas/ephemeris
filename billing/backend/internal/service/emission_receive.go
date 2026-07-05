package service

import (
	"fmt"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// EmissionReceive is responsible for handling the business logic of receiving emissions.
type EmissionReceive struct {
	Base
	receiver port.Issuer
}

// NewEmissionReceive creates a new instance of EmissionReceive.
func NewEmissionReceive(repo port.Repository, logger port.Logger, issuer port.Issuer) *EmissionReceive {
	return &EmissionReceive{
		Base:     *NewBase(repo, logger),
		receiver: issuer,
	}
}

// Run processes the request to receive emissions and returns the response.
func (s *EmissionReceive) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(1, "Running EmissionReceive service")
	receiveDTO, ok := inDTO.(*dto.EmissionReceiveRequest)
	if !ok {
		s.logger.IPrintf(1, "Invalid input type")
		return dto.NewEmissionReceiveResponse(400, "error", "Invalid input type", 0, 0, 0)
	}
	// Validate the input DTO
	if err := receiveDTO.Validate(s.repo); err != nil {
		s.logger.IPrintf(1, "Validation error: %v", err)
		return dto.NewEmissionReceiveResponse(400, "error", err.Error(), 0, 0, 0)
	}
	// Call the receiver to process the emission receive
	fileItems, err := s.receiver.ReceiveEmission(receiveDTO.Source)
	if err != nil {
		s.logger.IPrintf(1, "Error receiving emission: %v", err)
		return dto.NewEmissionReceiveResponse(400, "error", err.Error(), 0, 0, 0)
	}
	// Get the domain emission from the DTO
	emission := receiveDTO.GetDomain()
	// Merge the received emission items with the existing emission items
	if err := s.mergeEmissionItems(emission, fileItems); err != nil {
		s.logger.IPrintf(1, "Error merging emission items: %v", err)
		return dto.NewEmissionReceiveResponse(400, "error", err.Error(), 0, 0, 0)
	}
	// Process the emission receive logic here
	s.logger.IPrintf(1, "Emission received successfully: ID %d, Quantity %d, Amount %.2f",
		emission.ID, emission.Quantity, emission.Amount)
	// Return a successful response
	return dto.NewEmissionReceiveResponse(200, "success", "Emission received successfully",
		emission.ID, emission.Quantity, emission.Amount)
}

// mergeEmissionItems merges the received emission items with the existing emission items in the domain model.
func (s *EmissionReceive) mergeEmissionItems(emission *domain.Emission,
	fileItems map[int64]domain.EmissionItem) error {
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
	}
	// Save the updated emission to the repository
	if err := s.repo.Save(emission); err != nil {
		return fmt.Errorf("failed to save updated emission: %v", err)
	}
	return nil
}
