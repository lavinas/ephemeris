package driver

import (
	"flag"
	"fmt"

	"billing/internal/service"
)

// CreateCustomerCsv is responsible for handling the business logic of creating customers from a CSV file.
type CreateCustomerCsv struct {
	service service.CreateCustomerService
}

// NewCreateCustomerCsv creates a new instance of CreateCustomerCsv.
func NewCreateCustomerCsv(service service.CreateCustomerService) *CreateCustomerCsv {
	return &CreateCustomerCsv{
		service: service,
	}
}

// Run processes the CSV file and creates customers based on the data.
func (c *CreateCustomerCsv) Run() error {
	filePath := flag.String("file", "", "Path to the CSV file")
	flag.Parse()
	if *filePath == "" {
		return fmt.Errorf("file path is required")
	}
	return c.service.RunCSV(*filePath)
}
