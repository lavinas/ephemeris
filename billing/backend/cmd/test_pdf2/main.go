package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
	"time"
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
	// logoBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(logoData)
	logoBase64 := base64.StdEncoding.EncodeToString(logoData)
	// customerDoc := "044.123.456-78"
	customerEmail := "lavinas@gmail.com"
	// paymentDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	// Sample data for the receipt
	data := dto.IssuerData{
		VendorLogoBase64:   template.URL(logoBase64),
		VendorName:         "Estúdio Amelia Cardoso",
		VendorDocument:     "27.928.875/0001-04",
		VendorEmail:        "financeiro@ameliacardoso.com.br",
		VendorWhatsApp:     "(11) 98088-8399",
		VendorSMTPHost:     "smtp.zoho.com",
		VendorSMTPPort:     587,
		VendorSMTPUsername: "financeiro@ameliacardoso.com.br",
		VendorSMTPPassword: "pwd22Adm**",
		InvoiceNumber:      56766,
		CustomerName:       "Paulo Celso Lavinas Barbosa",
		CustomerNickname:   "paulo_lavinas",
		CustomerDocument:   nil,
		CustomerEmail:      &customerEmail,
		InvoiceDate:        time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		DueDate:            time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
		PaymentDate:        nil,
		Items: []dto.ReceiptItem{
			{Description: "aulas de canto de 60 minutos em junho de 2026", Quantity: 2, UnitPrice: 300.0, Total: 600.0},
			{Description: "aulas de piano de 60 minutos em junho de 2026", Quantity: 1, UnitPrice: 20.0, Total: 20.0},
		},
		TotalAmount: 620.0,
	}

	receipter := driven.NewIssuer()

	html_pdf, err := os.ReadFile("./templates/receipt_pdf.html")
	if err != nil {
		fmt.Println("Error reading HTML template:", err)
		return
	}
	htmlPDFContent := string(html_pdf)

	html_email, err := os.ReadFile("./templates/receipt_email.html")
	if err != nil {
		fmt.Println("Error reading email HTML template:", err)
		return
	}
	htmlEmailContent := string(html_email)

	// Generate the PDF receipt
	pdf, err := receipter.GetBase64(&data, htmlPDFContent)
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

	// Send the receipt via email
	err = receipter.SendMail(&data, htmlPDFContent, htmlEmailContent)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}

	fmt.Println("PDF receipt generated successfully: receipt.pdf")

}
