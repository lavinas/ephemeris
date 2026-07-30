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

const (
	qrcode    = "iVBORw0KGgoAAAANSUhEUgAAAQAAAAEAAQMAAABmvDolAAAABlBMVEX///8AAABVwtN+AAADWklEQVR42uyYMW7zOhCEh1DBkjcwLyJY13JhgAJS+FoydBH6BixZCJqHWdpJ3sMr/uK3wiIEAtjM54Lk7uzs4nf9rr+9JpKLK9wAJOaA4c6C4UFy7wY4A359lDossVyGHGraYyAfxf51EBB5W/V1yuG25oArUC7To3DrDNgD4HK4bY64OoZ5dEcDC0JNd/JGlnqdcpjRGQD4xbHdpGKP92xB+D0e3gwo7BVy13MMt28f/pUXPw1oDRnkvSWp/9gR5v+IyJsBy03q3mLxZHvEy8DiuXQDuMw6noCry0h6bjvF6EpN+2EAkVYWkiyXMRJXxFDHSP+8zh6AM0IdSL9NOVQAnjsK4AqG/SjAtCsWP7scuJI13XNJS6RnP8BEACdQIgtEHYCcx1OonwJyABDIDKQFBaOjyILpUfzWD6ACNDyor+HGPdRE/USyfxgA7ZO0IFe0b4hhHmNfQJRbILdzhPwV0l3ZKse1HAVIu9YMRTur8jOR+oBXWewEuCgLEjP8mk1Awjw9+CUgbwcmlrQ5FRt5iB3VEnRE4Gde/DxgSqvc3CMucPQflJplvELuEIBkDnXYYSnpaR6erOgHmCQgfAqImZx7DvOgU+AowOXWPugm0/owfYDno9S0dAO0UxS/AcHc8rDIkX499wEA9Jq7gjyWRBaJmL0mP21pF8D89DeUf+BsBTzylZsHAI4W9l7+Icm2XBGL35qH7wVolchemeqaOZ+lJICfz0cBirTtZPpQ/Epqv/jtFF5q3wOgvJhYJLCUwMqxm6P4cqRvBxxLsj4UkTfuqANz60yHfgDIiJ4g69461lk71gweBsgVL8rESWEv/dyjqT0/ln4A+SuqUY2tNvEu2xNZX+PH9wOI5til9k1Ir5I1nILNGzsBzjbxCH6eaLbQS2kviIXz+ShArniz1i/S2uePJQbVIN8T8Gz9tD+PCLBGbIvlZbQOACbCL4BCTq9ps7Jio7TPvPh5wKZqu/3xtjj+j8l5O9AG+0Ta0fQzLXI1383ezwM2p402/0Ay8s52irQfBkTeFvXpKj2t2SK89fLnvgAbqqsa7q2t8GubihwJSLuczLAcl3qc1yX3ArTnloghbZG4qoi3+ngU8BzsW25qVZUd88hDP8Dv+l1/vv4JAAD//zdiQXZVEFCPAAAAAElFTkSuQmCC"
	copyPaste = "00020101021126810014BR.GOV.BCB.PIX0114279288750001040241Aula de canto de 30 minutos em 09/07/202652040000530398654041.505802BR5925BARBOSA E CARDOSO PREPARA6009SAO PAULO62100506123456630463D7"
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
	customerDoc := "044.123.456-78"
	customerEmail := "lavinas@gmail.com"
	data := dto.IssuerData{
		VendorLogoBase64:     template.URL(logoBase64),
		VendorName:           "Estúdio Amelia Cardoso",
		VendorDocument:       "27.928.875/0001-04",
		VendorEmail:          "financeiro@ameliacardoso.com.br",
		VendorWhatsApp:       "(11) 98088-8399",
		VendorSMTPHost:       "smtp.zoho.com",
		VendorSMTPPort:       587,
		VendorSMTPUsername:   "financeiro@ameliacardoso.com.br",
		VendorSMTPPassword:   "pwd22Adm**",
		VendorPixQRBase64:    template.URL("data:image/png;base64," + qrcode),
		VendorPixCopyPaste:   copyPaste,
		VendorPixName:        "Estúdio Amelia Cardoso",
		VendorBank:           "Banco do Brasil",
		VendorAgency:         "1234-5",
		VendorAccount:        "67890-1",
		InvoiceNumber:        56766,
		CustomerFirstName:    "Paulo",
		CustomerName:         "Paulo Celso Lavinas Barbosa",
		CustomerNickname:     "paulo_lavinas",
		CustomerDocumentType: "CPF",
		CustomerDocument:     customerDoc,
		CustomerEmail:        customerEmail,
		InvoiceDate:          time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		InvoiceDueDate:       time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
		InvoicePaymentDate:   time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
		InvoiceItems: []dto.ReceiptItem{
			{Description: "aulas de canto de 60 minutos em junho de 2026", Quantity: 2, UnitPrice: 300.0, Total: 600.0},
			{Description: "aulas de piano de 60 minutos em junho de 2026", Quantity: 1, UnitPrice: 20.0, Total: 20.0},
		},
		InvoiceTotalAmount: 620.0,
	}

	receipter := driven.NewIssuer()
	err = sendEmail(receipter, &data, "Estúdio Amelia Cardoso - fatura", "./templates/invoice_pdf.html", "./templates/invoice_email.html")
	if err != nil {
		fmt.Println("Error generating and sending PDF invoice:", err)
		return
	}
	err = sendEmail(receipter, &data, "Estúdio Amelia Cardoso - recibo", "./templates/receipt_pdf.html", "./templates/receipt_email.html")
	if err != nil {
		fmt.Println("Error generating and sending PDF receipt:", err)
		return
	}
	fmt.Println("PDF receipt generated successfully")

}

// writeAndSendEmail generates a PDF receipt and sends it via email using the provided Issuer component.
func sendEmail(receipter *driven.Issuer, data *dto.IssuerData, subject, pdfPath, emailPath string) error {
	html_pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("error reading HTML template: %v", err)
	}
	htmlPDFContent := string(html_pdf)

	html_email, err := os.ReadFile(emailPath)
	if err != nil {
		return fmt.Errorf("error reading email HTML template: %v", err)
	}
	htmlEmailContent := string(html_email)

	// Send the receipt via email
	err = receipter.SendMail(data, subject, htmlPDFContent, htmlEmailContent)
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}
	return nil
}
