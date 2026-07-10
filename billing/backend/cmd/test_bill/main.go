package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
	"time"
)

func main() {
	logger, _ := driven.NewSimpleLogger("stdout", 2)
	pdfGenerator := driven.NewBiller(logger, "./files/bill/output.pdf")

	notes := make([]string, 0)
	notes = append(notes, "Seguem os dados para o depósito ou PIX (CNPJ: 27.928.875/0001-04)")
	notes = append(notes, "")
	notes = append(notes, "Por favor, assim que efetivar o depósito queira nos enviar o comprovante para o email financeiro@ameliacardoso.com.br, para que o horário da aula seja confirmado e que seja providenciado o recibo.")
	notes = append(notes, "")
	notes = append(notes, "Santander :")
	notes = append(notes, "Ag: 0985")
	notes = append(notes, "CC 13001001-4")
	notes = append(notes, "CNPJ: 27.928.875/0001-04   ")
	notes = append(notes, "Razão Social: BARBOSA E CARDOSO PREPARAÇÃO VOCAL E PRODUÇÕES MUSICAIS LTDA")
	notes = append(notes, "")
	notes = append(notes, "Att")
	notes = append(notes, "Estudio de aulas Amélia Cardoso")

	request := dto.BillerRequest{
		InvoiceID:   "123.456",
		InvoiceDate: time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
		InvoiceDue:  time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC), // Due in 30 days
		Vendor: dto.BillerVendor{
			Logo:     "./images/logo_amelia.png",
			Name:     "Estudio Amelia Cardoso",
			Document: "27.928.875/0001-04",
			Address:  strPtr("123 Main St"),
			Postcode: strPtr("12345"),
			City:     strPtr("Cityville"),
			State:    strPtr("State"),
			Country:  strPtr("Country"),
			Email:    strPtr("financeiro@ameliacardoso.com.br"),
			Whatsapp: strPtr("(11) 98088-8399"),
			PixKey:   strPtr("1234567890"),
		},
		Customer: dto.BillerCustomer{
			Name:     "Rui Miranda Ribeiro Facó",
			Document: strPtr("044.123.456-78"),
			Email:    strPtr("rui.miranda@example.com"),
			Whatsapp: strPtr("(11) 91234-5678"),
		},
		Items: []dto.BillerItem{
			{Description: "Item 1", Quantity: 2, Price: 10.0},
			{Description: "Item 2", Quantity: 1, Price: 20.0},
		},
		Notes: &notes,
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
