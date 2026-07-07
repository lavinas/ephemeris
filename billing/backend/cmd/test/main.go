package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
)

func main() {
	logger, _ := driven.NewSimpleLogger("stdout", 2)
	pdfGenerator := driven.NewBiller(logger, "./files/bill/output.pdf")

	request := dto.BillerRequest{
		Vendor: dto.BillerVendor{
			Logo:     "./images/logo_amelia.png",
			Name:     "Estudio Amelia Cardoso",
			Address:  strPtr("123 Main St"),
			Postcode: strPtr("12345"),
			City:     strPtr("Cityville"),
			State:    strPtr("State"),
			Country:  strPtr("Country"),
			Email:    strPtr("info@estudioameliacardoso.com"),
			Whatsapp: strPtr("+1234567890"),
			PixKey:   strPtr("1234567890"),
		},
		Customer: dto.BillerCustomer{
			Name:     "John Doe",
			Document: strPtr("123456789"),
			Email:    strPtr("john.doe@example.com"),
			Whatsapp: strPtr("+1234567890"),
		},
		Items: []dto.BillerItem{
			{Description: "Item 1", Quantity: 2, Price: 10.0},
			{Description: "Item 2", Quantity: 1, Price: 20.0},
		},
		Notes: strPtr("Thank you for your business!"),
	}

	err := pdfGenerator.GeneratePDF(request)
	if err != nil {
		logger.IPrintf(1, "Error generating PDF: %v", err)
	}
}

// strPtr is a helper function to create a pointer to a string.
func strPtr(s string) *string {
	return &s
}
