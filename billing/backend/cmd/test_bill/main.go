package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
	"time"
)

const (
	qrcode    = "iVBORw0KGgoAAAANSUhEUgAAAQAAAAEAAQMAAABmvDolAAAABlBMVEX///8AAABVwtN+AAADWklEQVR42uyYMW7zOhCEh1DBkjcwLyJY13JhgAJS+FoydBH6BixZCJqHWdpJ3sMr/uK3wiIEAtjM54Lk7uzs4nf9rr+9JpKLK9wAJOaA4c6C4UFy7wY4A359lDossVyGHGraYyAfxf51EBB5W/V1yuG25oArUC7To3DrDNgD4HK4bY64OoZ5dEcDC0JNd/JGlnqdcpjRGQD4xbHdpGKP92xB+D0e3gwo7BVy13MMt28f/pUXPw1oDRnkvSWp/9gR5v+IyJsBy03q3mLxZHvEy8DiuXQDuMw6noCry0h6bjvF6EpN+2EAkVYWkiyXMRJXxFDHSP+8zh6AM0IdSL9NOVQAnjsK4AqG/SjAtCsWP7scuJI13XNJS6RnP8BEACdQIgtEHYCcx1OonwJyABDIDKQFBaOjyILpUfzWD6ACNDyor+HGPdRE/USyfxgA7ZO0IFe0b4hhHmNfQJRbILdzhPwV0l3ZKse1HAVIu9YMRTur8jOR+oBXWewEuCgLEjP8mk1Awjw9+CUgbwcmlrQ5FRt5iB3VEnRE4Gde/DxgSqvc3CMucPQflJplvELuEIBkDnXYYSnpaR6erOgHmCQgfAqImZx7DvOgU+AowOXWPugm0/owfYDno9S0dAO0UxS/AcHc8rDIkX499wEA9Jq7gjyWRBaJmL0mP21pF8D89DeUf+BsBTzylZsHAI4W9l7+Icm2XBGL35qH7wVolchemeqaOZ+lJICfz0cBirTtZPpQ/Epqv/jtFF5q3wOgvJhYJLCUwMqxm6P4cqRvBxxLsj4UkTfuqANz60yHfgDIiJ4g69461lk71gweBsgVL8rESWEv/dyjqT0/ln4A+SuqUY2tNvEu2xNZX+PH9wOI5til9k1Ir5I1nILNGzsBzjbxCH6eaLbQS2kviIXz+ShArniz1i/S2uePJQbVIN8T8Gz9tD+PCLBGbIvlZbQOACbCL4BCTq9ps7Jio7TPvPh5wKZqu/3xtjj+j8l5O9AG+0Ta0fQzLXI1383ezwM2p402/0Ay8s52irQfBkTeFvXpKj2t2SK89fLnvgAbqqsa7q2t8GubihwJSLuczLAcl3qc1yX3ArTnloghbZG4qoi3+ngU8BzsW25qVZUd88hDP8Dv+l1/vv4JAAD//zdiQXZVEFCPAAAAAElFTkSuQmCC"
	copyPaste = "00020101021126810014BR.GOV.BCB.PIX0114279288750001040241Aula de canto de 30 minutos em 09/07/202652040000530398654041.505802BR5925BARBOSA E CARDOSO PREPARA6009SAO PAULO62100506123456630463D7"
)

func main() {
	logger, _ := driven.NewSimpleLogger("stdout", 2)
	pdfGenerator := driven.NewBillerMaroto(logger)

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

	request := &dto.BillerRequest{
		InvoiceID:   123456789,
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
		},
		Customer: dto.BillerCustomer{
			Name:     "Paulo Barbosa",
			Document: strPtr("044.123.456-78"),
			Email:    strPtr("lavinas@gmail.com"),
			Whatsapp: strPtr("(11) 98087-6112"),
		},
		Items: []dto.BillerItem{
			{Description: "aulas de canto de 60 minutos em junho de 2026", Quantity: 2, Price: 300.0},
			{Description: "aulas de piano de 60 minutos em junho de 2026", Quantity: 1, Price: 20.0},
		},
		Receive: dto.BillerReceive{
			BankAccount: &dto.BillerBankAccount{
				BankName:         "Santander (033)",
				BankAgency:       "0985",
				BankAccount:      "13001001-4",
				ReceiverName:     "BARBOSA E CARDOSO PREPARAÇÃO VOCAL E PRODUÇÕES MUSICAIS LTDA",
				ReceiverDocument: "27.928.875/0001-04",
			},
			Pix: &dto.BillerPix{
				PixKey:       "27.928.875/0001-04",
				ReceiverName: "Estudio Vocal Amelia Cardoso",
				PixCopyPaste: copyPaste,
				PixQRCode:    qrcode,
			},
		},
	}

	err := pdfGenerator.SendMail(request)
	if err != nil {
		logger.IPrintf(1, "Error generating PDF: %v", err)
	}

}

// strPtr is a helper function to create a pointer to a string.
func strPtr(s string) *string {
	return &s
}
