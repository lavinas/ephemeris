package service

import (
	"time"
	"strings"
	"encoding/base64"
	"fmt"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

const (
	filePath    = "./files/"
	filePattern = "nfe_<yyyy>_<mm>_<id>.txt"
)

// TaxGenerate is responsible for handling the business logic of sending emissions.
type TaxGenerate struct {
	Base
	sender port.Taxer
}

// NewTaxGenerate creates a new instance of TaxGenerate.
func NewTaxGenerate(repo port.Repository, logger port.Logger, taxer port.Taxer) *TaxGenerate {
	return &TaxGenerate{
		Base:   *NewBase(repo, logger),
		sender: taxer,
	}
}

// Run processes the request to send emissions and returns the response.
func (s *TaxGenerate) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(1, "Running TaxGenerate service")
	sendDTO, ok := inDTO.(*dto.TaxGenerateRequest)
	if !ok {
		s.logger.IPrintf(1, "Invalid input type")
		return dto.NewTaxGenerateResponse(400, "error", "Invalid input type", 0, 0, 0.0, "", "")
	}
	if err := sendDTO.Validate(s.repo); err != nil {
		s.logger.IPrintf(1, "Validation error: %v", err)
		return dto.NewTaxGenerateResponse(400, "error", err.Error(), 0, 0, 0.0, "", "")
	}
	vendor, err := s.getVendor(sendDTO.Vendor)
	if err != nil {
		return dto.NewTaxGenerateResponse(500, "error", "Failed to retrieve vendor", 0, 0, 0.0, "", "")
	}
	lastRPS, err := s.getLastRSPNumber(vendor.ID, vendor.LastRps)
	if err != nil {
		return dto.NewTaxGenerateResponse(500, "error", "Failed to retrieve last RPS number", 0, 0, 0.0, "", "")
	}
	domainEmission, err := s.getEmission(vendor, sendDTO.EmissionDate,
		sendDTO.InvoiceStartDate, sendDTO.InvoiceEndDate, lastRPS)
	if err != nil {
		return dto.NewTaxGenerateResponse(500, "error", "Failed to retrieve emission", 0, 0, 0.0, "", "")
	}
	doc, name, err := s.SendAndSave(domainEmission)
	if err != nil {
		s.logger.IPrintf(1, "Failed to send and save emission: %v", err)
		return dto.NewTaxGenerateResponse(500, "error", "Failed to send and save emission", 0, 0, 0.0, "", "")
	}
	s.logger.IPrintf(1, "Emission sent successfully: ID %d, Quantity %d, Amount %.2f",
		domainEmission.ID, domainEmission.Quantity, domainEmission.Amount)
	// Return a successful response
	return dto.NewTaxGenerateResponse(200, "success", "Emission sent successfully",
		domainEmission.ID, domainEmission.Quantity, domainEmission.Amount, *doc, *name)
}

// getEmission retrieves the emission data for the specified vendor and date range from the repository.
func (s *TaxGenerate) getEmission(vendor *domain.Vendor, emissionDate, startDate,
	endDate string, lastRPS int64) (*domain.Emission, error) {
	s.logger.IPrintf(2, "Retrieving emission for vendor: %s, start date: %s, end date: %s, emission date: %s",
		vendor.Nickname, startDate, endDate, emissionDate)
	sd, _ := time.Parse("2006-01-02", startDate)
	ed, _ := time.Parse("2006-01-02", endDate)
	emissionDateParsed, _ := time.Parse("2006-01-02", emissionDate)
	invs, err := s.repo.GetInvoicesByPeriod(vendor.ID, sd, ed)
	if err != nil {
		s.logger.IPrintf(1, "Failed to retrieve invoices: %v", err)
		return nil, err
	}
	emissionItems, totalAmount, err := s.getEmissionItems(invs, lastRPS)
	if err != nil {
		s.logger.IPrintf(1, "Failed to retrieve emission items: %v", err)
		return nil, err
	}
	quantity := len(emissionItems)
	emission := domain.NewEmission(vendor, sd, ed, emissionDateParsed,
		lastRPS+1, lastRPS+int64(quantity), totalAmount, quantity, emissionItems)
	s.logger.IPrintf(2, "Emission created: ID %d, Quantity %d, Amount %.2f", 
		emission.ID, emission.Quantity, emission.Amount)
	return emission, nil
}

// getEmissionItems retrieves the emission items for the specified invoices and RPS range.
func (s *TaxGenerate) getEmissionItems(invoices []domain.Invoice, lastRPS int64) ([]*domain.EmissionItem, float64, error) {
	emissionItems := make([]*domain.EmissionItem, 0, len(invoices))
	rps := lastRPS
	totalAmount := 0.0
	for _, inv := range invoices {
		if inv.IsTaxable() == false {
			s.logger.IPrintf(3, "Skipping invoice ID: %d as it is not taxable", inv.ID)
			continue
		}
		s.logger.IPrintf(3, "Invoice ID: %d, Amount: %.2f", inv.ID, inv.Amount)
		totalAmount += inv.Amount
		rps++
		emissionItems = append(emissionItems, domain.NewEmissionItem(0, inv.ID, rps, inv))
	}
	return emissionItems, totalAmount, nil
}

// getVendor retrieves the vendor information for the specified vendor ID from the repository.
func (s *TaxGenerate) getVendor(nickname string) (*domain.Vendor, error) {
	vendor, err := s.repo.GetVendor(nickname)
	if err != nil {
		s.logger.IPrintf(1, "Failed to retrieve vendor: %v", err)
		return nil, err
	}
	return vendor, nil
}

// getLastRSPNumber retrieves the last RPS number for the specified vendor from the repository.
func (s *TaxGenerate) getLastRSPNumber(vendorID int64, defaultRPS int64) (int64, error) {
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
func (s *TaxGenerate) SendAndSave(emission *domain.Emission) (*string, *string, error) {
	s.logger.IPrintf(2, "Sending emission ID: %d, Quantity: %d, Amount: %.2f", emission.ID, emission.Quantity, emission.Amount)
	if err := s.repo.BeginTransaction(); err != nil {
		return nil, nil, err
	}
	defer s.repo.RollbackTransaction()
	doc, name, err := s.getDocument(emission)
	if err != nil {
		return nil, nil, err
	}
	if err := s.repo.Save(emission); err != nil {
		return nil, nil, err
	}
	if err := s.saveInvoices(emission); err != nil {
		return nil, nil, err
	}
	if err := s.repo.CommitTransaction(); err != nil {
		return nil, nil, err
	}
	s.logger.IPrintf(2, "Emission ID: %d sent and saved successfully", emission.ID)
	return doc, name, nil
}

// saveInvoices updates the invoices associated with the emission in the repository.
func (s *TaxGenerate) saveInvoices(emission *domain.Emission) error {
	invoices := make([]domain.Invoice, 0, len(emission.EmissionItems))
	for _, item := range emission.EmissionItems {
		invoices = append(invoices, item.Invoice)
	}
	return s.repo.Save(invoices)
}

// getEmission retrieves the emission data for the specified vendor and date range from the repository.
func (s *TaxGenerate) getDocument(emission *domain.Emission) (*string, *string, error) {
	var tax strings.Builder
	if err := s.sender.GetEmission(emission, &tax); err != nil {
		return nil, nil, err
	}
	taxStr := base64.StdEncoding.EncodeToString([]byte(tax.String()))

	name := "nfe_%s_%s_%d.txt"
	name = fmt.Sprintf(name, emission.Vendor.Nickname, emission.EmissionDate.Format("2006-01"), emission.ID)

	return &taxStr, &name, nil
}
