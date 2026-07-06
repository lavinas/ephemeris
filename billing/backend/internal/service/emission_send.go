package service

import (
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
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
	s.logger.IPrintf(1, "Running EmissionSend service")
	sendDTO, ok := inDTO.(*dto.EmissionSendRequest)
	if !ok {
		s.logger.IPrintf(1, "Invalid input type")
		return dto.NewEmissionSendResponse(400, "error", "Invalid input type", 0, 0, 0)
	}
	if err := sendDTO.Validate(s.repo); err != nil {
		s.logger.IPrintf(1, "Validation error: %v", err)
		return dto.NewEmissionSendResponse(400, "error", err.Error(), 0, 0, 0)
	}
	vendor, err := s.getVendor(sendDTO.Vendor)
	if err != nil {
		return dto.NewEmissionSendResponse(500, "error", "Failed to retrieve vendor", 0, 0, 0)
	}
	lastRPS, err := s.getLastRSPNumber(vendor.ID, vendor.LastRps)
	if err != nil {
		return dto.NewEmissionSendResponse(500, "error", "Failed to retrieve last RPS number", 0, 0, 0)
	}
	domainEmission, invoices, err := s.getEmission(vendor, sendDTO.EmissionDate,
		sendDTO.InvoiceStartDate, sendDTO.InvoiceEndDate, lastRPS)
	if err != nil {
		return dto.NewEmissionSendResponse(500, "error", "Failed to retrieve emission", 0, 0, 0)
	}
	if err := s.SendAndSave(domainEmission, invoices); err != nil {
		s.logger.IPrintf(1, "Failed to send and save emission: %v", err)
		return dto.NewEmissionSendResponse(500, "error", "Failed to send and save emission", 0, 0, 0)
	}
	s.logger.IPrintf(1, "Emission sent successfully: ID %d, Quantity %d, Amount %.2f",
		domainEmission.ID, domainEmission.Quantity, domainEmission.Amount)
	// Return a successful response
	return dto.NewEmissionSendResponse(200, "success", "Emission sent successfully",
		domainEmission.ID, domainEmission.Quantity, domainEmission.Amount)
}

// getEmission retrieves the emission data for the specified vendor and date range from the repository.
func (s *EmissionSend) getEmission(vendor *domain.Vendor, emissionDate, startDate,
	endDate string, lastRPS int64) (*domain.Emission, []domain.Invoice, error) {
	s.logger.IPrintf(2, "Retrieving emission for vendor: %s, start date: %s, end date: %s, emission date: %s",
		vendor.Nickname, startDate, endDate, emissionDate)
	sd, _ := time.Parse("2006-01-02", startDate)
	ed, _ := time.Parse("2006-01-02", endDate)
	emissionDateParsed, _ := time.Parse("2006-01-02", emissionDate)

	invs, err := s.repo.GetInvoicesByPeriod(vendor.ID, sd, ed)
	if err != nil {
		s.logger.IPrintf(1, "Failed to retrieve invoices: %v", err)
		return nil, nil, err
	}
	invsOut := make([]domain.Invoice, 0)
	totalAmount := 0.0
	quantity := 0
	emissionItems := make([]*domain.EmissionItem, 0)
	rps := lastRPS
	for _, inv := range invs {
		if inv.IsTaxable() == false {
			s.logger.IPrintf(3, "Skipping invoice ID: %d as it is not taxable", inv.ID)
			continue
		}
		s.logger.IPrintf(3, "Invoice ID: %d, Amount: %.2f", inv.ID, inv.Amount)
		totalAmount += inv.Amount
		quantity++
		rps++
		emissionItems = append(emissionItems, domain.NewEmissionItem(0, inv.ID, rps, inv))
		inv.UpdatedAt = time.Now()
		now := time.Now()
		inv.TaxDate = &now
		invsOut = append(invsOut, inv)
	}
	emission := domain.NewEmission(vendor, sd, ed, emissionDateParsed,
		lastRPS+1, rps-1, totalAmount, quantity, emissionItems)
	s.logger.IPrintf(2, "Emission created: ID %d, Quantity %d, Amount %.2f", emission.ID, emission.Quantity, emission.Amount)
	return emission, invsOut, nil
}

// getVendor retrieves the vendor information for the specified vendor ID from the repository.
func (s *EmissionSend) getVendor(nickname string) (*domain.Vendor, error) {
	vendor, err := s.repo.GetVendor(nickname)
	if err != nil {
		s.logger.IPrintf(1, "Failed to retrieve vendor: %v", err)
		return nil, err
	}
	return vendor, nil
}

// getLastRSPNumber retrieves the last RPS number for the specified vendor from the repository.
func (s *EmissionSend) getLastRSPNumber(vendorID int64, defaultRPS int64) (int64, error) {
	lastRPS, err := s.repo.GetEmissionLastRPS(vendorID)
	if err != nil {
		s.logger.IPrintf(1, "Failed to retrieve last RPS number: %v", err)
		return 0, err
	}
	if lastRPS == 0 {
		return defaultRPS, nil
	}
	return lastRPS, nil
}

// saveAll saves the emission and updates the invoices in the repository.
func (s *EmissionSend) SendAndSave(emission *domain.Emission, invoices []domain.Invoice) error {
	s.logger.IPrintf(2, "Sending emission ID: %d, Quantity: %d, Amount: %.2f", emission.ID, emission.Quantity, emission.Amount)
	if err := s.repo.BeginTransaction(); err != nil {
		return err
	}
	defer s.repo.RollbackTransaction()
	if err := s.sender.SendEmission(emission); err != nil {
		return err
	}
	if err := s.repo.Save(emission); err != nil {
		return err
	}
	if err := s.repo.Save(invoices); err != nil {
		return err
	}
	if err := s.repo.CommitTransaction(); err != nil {
		return err
	}
	s.logger.IPrintf(2, "Emission ID: %d sent and saved successfully", emission.ID)
	return nil
}
