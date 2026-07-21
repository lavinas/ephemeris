package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
)

// This program generates a PDF receipt using the Receipter component.
func main() {
	logo, err := os.Open("./images/logo_amelia.png")
	if err != nil {
		fmt.Println("Error opening logo file:", err)
		return
	}
	defer logo.Close()

	logoData, err := io.ReadAll(logo)
	if err != nil {
		fmt.Println("Error reading logo file:", err)
		return
	}
	logoBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(logoData)

	// Sample data for the receipt
	data := dto.ReceiptData{
		VendorLogoBase64: template.URL(logoBase64),
		VendorName:       "Estúdio Amelia Cardoso",
		VendorDocument:   "27.928.875/0001-04",
		VendorEmail:      "financeiro@amelicardoso.com.br",
		VendorWhatsApp:   "(11) 98088-8399",
		InvoiceNumber:    "56766",
		CustomerName:     "Paulo Celso Lavinas Barbosa",
		CustomerDocument: "044.123.456-78",
		CustomerEmail:    "lavinas@gmail.com",
		IssueDate:        "2026-06-01",
		DueDate:          "2026-06-30",
		PaymentDate:      "2026-06-15",
		Items: []dto.ReceiptItem{
			{Description: "aulas de canto de 60 minutos em junho de 2026", Quantity: 2, UnitPrice: 300.0, Total: 600.0},
			{Description: "aulas de piano de 60 minutos em junho de 2026", Quantity: 1, UnitPrice: 20.0, Total: 20.0},
		},
		TotalAmount: 620.0,
	}

	receipter, err := driven.NewReceipter("./templates/pdf/receipt.html")
	if err != nil {
		fmt.Println("Error creating receipter:", err)
		return
	}

	// Generate the PDF receipt
	pdf, err := receipter.GetReceiptBase64(data)
	if err != nil {
		fmt.Println("Error generating PDF receipt:", err)
		return
	}

	// Save the PDF to a file
	err = os.WriteFile("./receipt.pdf", pdf, 0644)
	if err != nil {
		fmt.Println("Error saving PDF receipt:", err)
		return
	}

	fmt.Println("PDF receipt generated successfully: receipt.pdf")

}
