package service

import (
	"fmt"
	"io"
	"os"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// CreateCustomerService is responsible for handling the business logic of creating customers.
type CreateCustomerService struct {
	Base
}

// NewCreateCustomerService creates a new instance of CreateCustomerService.
func NewCreateCustomerService(repo port.Repository, logger port.Logger) *CreateCustomerService {
	return &CreateCustomerService{
		Base: *NewBase(repo, logger),
	}
}

// CreateCustomer creates a new customer using the provided details and saves it to the repository.
func (s *CreateCustomerService) Run(request dto.CreateCustomerRequest) dto.CreateCustomerResponse {
	s.logger.IPrintf(1, "Creating customer: %s", request.Name)
	customer := domain.NewCustomer(request.Name, request.Nickname, request.Document, request.Email, request.Whatsapp)
	// validate input
	if err := request.Validate(); err != nil {
		s.logger.IPrintf(1, "Erros validation %v", err)
		return dto.CreateCustomerResponse{
			ResponseBase: dto.ResponseBase{
				HttpCode: 400,
				Status:   "Bad Request",
				Message:  "Failed to create customer",
			},
		}
	}
	// save customer
	if err := s.repo.Save(customer); err != nil {
		s.logger.IPrintf(1, "Error creating customer: %v", err)
		return dto.CreateCustomerResponse{
			ResponseBase: dto.ResponseBase{
				HttpCode: 500,
				Status:   "error",
				Message:  "Failed to create customer",
			},
		}
	}
	// return ok
	s.logger.IPrintf(1, "Customer created successfully")
	return dto.CreateCustomerResponse{
		ResponseBase: dto.ResponseBase{
			HttpCode: 200,
			Status:   "success",
			Message:  "Customer created successfully",
		},
		ID: customer.ID,
	}
}

// RunCSV processes the CSV file and creates customers based on the data.
func (s *CreateCustomerService) RunCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	for {
		var name, nickname, email, whatsapp, document string
		_, err := fmt.Fscanf(file, "%s,%s,%s,%s,%s\n", &name, &nickname, &email, &whatsapp, &document)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		request := dto.CreateCustomerRequest{
			Name:     name,
			Nickname: nickname,
			Email:    email,
			Whatsapp: whatsapp,
			Document: document,
		}
		response := s.Run(request)
		s.logger.IPrintf(1, "Create Customer Response: %+v", response)
	}
	return nil
}
