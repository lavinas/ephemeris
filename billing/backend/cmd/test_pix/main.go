package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
)

// Main function to test the Pix payment payload generation.
func main() {
	logger, _ := driven.NewSimpleLogger("stdout", 0)
	pixToken := driven.NewPixer(logger)

	request := &dto.PixRequest{
		Key:         "27928875000104",
		Description: "Aula de canto de 30 minutos em 09/07/2026",
		Name:        "Estudio Vocal Amelia Cardoso",
		City:        "SAO PAULO",
		Amount:      1.5,
		Txid:        "123456",
	}

	payload, qrCode, err := pixToken.Get(request)
	if err != nil {
		println("Error generating Pix payload:", err.Error())
		return
	}
	println("Generated Pix Payload:", payload)
	println("Generated Pix QR Code:", qrCode)
}
