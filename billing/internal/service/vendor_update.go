package service

import (
	"fmt"

	"billing/internal/dto"
	"billing/internal/port"
)

// VendorUpdate is responsible for handling the business logic of updating vendors.
type VendorUpdate struct {
	Base
}

// NewVendorUpdate creates a new instance of VendorUpdate with the provided repository and logger.
func NewVendorUpdate(repo port.Repository, logger port.Logger) *VendorUpdate {
	return &VendorUpdate{
		Base: *NewBase(repo, logger),
	}
}

// Run executes the vendor update process using the provided request data and returns a response.
func (s *VendorUpdate) Run(inDTO port.InDTO) port.OutDTO {
	request, ok := inDTO.(*dto.VendorUpdateRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type")
		return dto.NewVendorUpdateResponse(400, "error", "Invalid input type")
	}
	s.logger.IPrintf(2, "Processing update vendor request: %v", request)
	if err := request.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewVendorUpdateResponse(400, "error", fmt.Sprintf("Validation failed: %v", err))
	}
	domainVendor := request.GetDomain()
	if domainVendor == nil {
		s.logger.IPrintf(2, "Failed to convert request to domain model")
		return dto.NewVendorUpdateResponse(500, "error", "contact support")
	}
	err := s.repo.Save(domainVendor)
	if err != nil {
		s.logger.IPrintf(2, "Failed to save vendor: %v", err)
		return dto.NewVendorUpdateResponse(500, "error", "contact support")
	}
	s.logger.IPrintf(2, "Successfully updated vendor: %s", request.Nickname)
	return dto.NewVendorUpdateResponse(200, "success", "Vendor updated successfully")
}
