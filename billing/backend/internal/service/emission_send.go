package service

import (
	"time"

	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/domain"
)

// EmissionSend is responsible for handling the business logic of sending emissions.
type EmissionSend struct {
	Base
	sender port.Issuer
}

// NewEmissionSend creates a new instance of EmissionSend.
func NewEmissionSend(repo port.Repository, logger port.Logger, issuer port.Issuer) *EmissionSend {
	return &EmissionSend{
		Base:   *NewBase(repo, logger),
		sender: issuer,
	}
}

// Run processes the request to send emissions and returns the response.
func (s *EmissionSend) Run(inDTO port.InDTO) port.OutDTO {

	sendDTO, ok := inDTO.(*dto.EmissionSendRequest)
	if !ok {
		s.logger.IPrintf(1, "Invalid input type")
		return dto.NewEmissionSendResponse(400, "error", "Invalid input type", 0, 0, 0)
	}
	if err := sendDTO.Validate(s.repo); err != nil {
		s.logger.IPrintf(1, "Validation error: %v", err)
		return dto.NewEmissionSendResponse(400, "error", err.Error(), 0, 0, 0)
	}
	// Retrieve the emission data from the repository
	domainEmission, err := s.getEmission(sendDTO.VendorID, sendDTO.InvoiceStartDate, sendDTO.InvoiceEndDate)
	if err != nil {
		s.logger.IPrintf(1, "Failed to retrieve emission: %v", err)
		return dto.NewEmissionSendResponse(500, "error", "Failed to retrieve emission", 0, 0, 0)
	}
	// Send the emission using the configured sender
	if err := s.sender.SendEmission(domainEmission); err != nil {
		s.logger.IPrintf(1, "Failed to send emission: %v", err)
		return dto.NewEmissionSendResponse(500, "error", "Failed to send emission", 0, 0, 0)
	}

	// Return a successful response
	return dto.NewEmissionSendResponse(200, "success", "Emission sent successfully",
		domainEmission.ID, domainEmission.Quantity, domainEmission.Amount)
}


// getEmission retrieves the emission data for the specified vendor and date range from the repository.
func (s *EmissionSend) getEmission(vendorID int64, startDate, endDate string) (*domain.Emission, error) {
	sd, _ := time.Parse("2006-01-02", startDate)
	ed, _ := time.Parse("2006-01-02", endDate)

	invs, err := s.repo.GetInvoicesByPeriod(vendorID, sd, ed)
	if err != nil {
		s.logger.IPrintf(1, "Failed to retrieve invoices: %v", err)
		return nil, err
	}
	items := make([]domain.EmissionItem, 0)
	var totalAmount float64
	var quantity int
	for _, inv := range invs {
		item := domain.NewEmissionItem(0, inv.ID, inv.Amount)
		totalAmount += inv.Amount
		quantity++
	}


	return nil, nil // Placeholder implementation; replace with actual repository call to retrieve emission data.
}