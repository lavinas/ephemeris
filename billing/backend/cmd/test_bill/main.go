package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
	"time"
)

const (
	qrcode    = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAQAAAAEAAQMAAABmvDolAAAABlBMVEX///8AAABVwtN+AAADZ0lEQVR42uyZMa6rPBSEB1G49A7ijSDYVopIIFFkW0bZiNmBSxeI+TV2yL33f80rXoiL6+bxuF+B7HPmzDj4Xb/rX6+B5GO3JAPQBnD2sFO7ktyrAXrA+CaS3tnUkukG8M415j+dBDjePWwCnJ06wEy9s2lYI7e6gEdAusFhfATLeXcRXXM6sFvOJEYy4tbrkysDAPNYY2qXEI1vSC6Bd9/8qIc3A7ns15huvbP3bw8/++LDgFYbLHI7+IYpb+cQfojIm4Ee9s6VOjuJRN5JGP079dUAANWJhh7ReBeP3mRq97OAgTCP/G2Mox4mIF6HYA3rAYCIdrdJWzoBVsdtp2GH2XAW0ASMXGnIYFPntKUuvzl6swagdxFZsprAqdVo1LZuF5wIDEEiVlpSD/psO3UXmNnXA9AmuFjU3l/AeSFTq/48DZA+ZO3yYDnN3cXRO5qtrwZoiCsaJs3HqSVxG0KU40qtPwsYAsYNMPNCXLuLpjJhuDIdjrQCoHdMXRMhk3PtkB1hBC72EJATAOnDgxGtKm1Q2Q+Mozbx1tcEXIv/fFbaTNriUU8DAIlWxLjQ5gegbGcafTVAQyrRaFJbliZ12eRw3k8Dgs3WfVxYrOmtCVLUqBldC9CX6Jdt81jcshKZEkd/HpCzaGq9i5o6nJeQT/MrkFYAOKArahbNdgG3gfa+weJZcicAAMbHSjMvgZR10Zde4b5Z1s8DgyY1Y3YLsqbaUuRENvVnATk+7PLALmsXjhl9jMUqAIcrYDnBcdJIyr0p8jUW3w40xLjJaMkMP/JELhnHvFTu8wCcnVqJ/FCsO2ePXHKnAtE8nolG81FjMertccVRBVCivZmaYOldzLYwtcH+73rhnYBsA5WaKSNakqmdtK2vsv88MOT/QtrF1MGaWflebuelMG8HZPa2RkIWyE2peUfUmyPjVAEworsANwWNIrkuXoFXGDwBGJ5h4TYEe9do3pqQ7zq+TvPzQOkLfXOZRH+YnBOAfLuYXUvxeHLsEcMaday1APnm/1ISTf4JYF4YDQO+HOn7gXKpTi7lNHOlXRUf0FcF+EvxoPmySB5e/vkQkLMAF2WrJAvZFUtR45dlrQDQKas3lpAFhFPzvDX6Xg/vBZ6/qUlRizXVNJQ1Pe6CagB+1+/6+/VfAAAA//9tb2M83Ye9zwAAAABJRU5ErkJggg=="
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

	request := dto.BillerRequest{
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
			Name:     "Rui Miranda Ribeiro Facó",
			Document: strPtr("044.123.456-78"),
			Email:    strPtr("rui.miranda@example.com"),
			Whatsapp: strPtr("(11) 91234-5678"),
		},
		Items: []dto.BillerItem{
			{Description: "Item 1", Quantity: 2, Price: 10.0},
			{Description: "Item 2", Quantity: 1, Price: 20.0},
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

	err := pdfGenerator.Generate(request, "./files/bill/output.pdf")
	if err != nil {
		logger.IPrintf(1, "Error generating PDF: %v", err)
	}

	bin, err := pdfGenerator.Get(request)
	if err != nil {
		logger.IPrintf(1, "Error getting PDF binary: %v", err)
	}
	logger.IPrintf(1, "Tamanho binário: %d", len(bin))
}

// strPtr is a helper function to create a pointer to a string.
func strPtr(s string) *string {
	return &s
}
