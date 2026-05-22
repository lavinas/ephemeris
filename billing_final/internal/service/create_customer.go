package service

import (
	"fmt"
	"os"
	"encoding/csv"

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

// melhorar
// RunCSV processes the CSV file and creates customers based on the data.
func (s *CreateCustomerService) RunCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return err
	}
	if len(header) < 5 {
		return fmt.Errorf("invalid CSV header: %v", header)
	}
	if header[0] != "name" || header[1] != "nickname" || header[2] != "document" || header[3] != "email" || header[4] != "whatsapp" {
		return fmt.Errorf("invalid CSV header: %v", header)
	}
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	for _, record := range records {
		if len(record) < 5 {
			s.logger.IPrintf(1, "Skipping invalid record: %v", record)
			continue
		}
		name := record[0]
		nickname := record[1]
		document := record[2]
		email := record[3]
		whatsapp := record[4]
		request := dto.CreateCustomerRequest{
			Name:     name,
			Nickname: nickname,
			Document: document,
			Email:    email,
			Whatsapp: whatsapp,
		}
		response := s.Run(request)
		s.logger.IPrintf(1, "Create Customer Response: %+v", response)
	}
	return nil
}
