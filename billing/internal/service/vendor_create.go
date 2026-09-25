package service

import (
	"fmt"

	"billing/internal/dto"
	"billing/internal/port"
)

// VendorCreate is responsible for handling the business logic of creating vendors.
type VendorCreate struct {
	Base
}

// NewVendorCreate creates a new instance of VendorCreate.
func NewVendorCreate(repo port.Repository, logger port.Logger) *VendorCreate {
	return &VendorCreate{
		Base: *NewBase(repo, logger),
	}
}

// Run processes a vendor creation request and returns the response.
func (s *VendorCreate) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create vendor request: %v", inDTO)
	in, ok := inDTO.(*dto.VendorCreateRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected VendorCreateRequest")
		return dto.NewVendorCreateResponse(400, "bad request", "Invalid input type")
	}
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewVendorCreateResponse(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err))
	}
	if err := s.repo.Save(in.GetDomain()); err != nil {
		s.logger.IPrintf(2, "Failed to save vendor: %v", err)
		return dto.NewVendorCreateResponse(500, "internal error", "contact support please")
	}
	s.logger.IPrintf(2, "Successfully processed vendor creation request for %s", in.Nickname)
	return dto.NewVendorCreateResponse(200, "success", "Vendor created successfully")
}
