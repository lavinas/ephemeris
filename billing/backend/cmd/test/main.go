package main

import (
	"fmt"
	"billing/internal/dto"
)

func validarNumeroInternacional(numeroInput string) {

	// Passando "" no segundo parâmetro, a lib descobre o país pelo "+"
	num, err := dto.ValidateCellNumber(numeroInput)
	if err == nil {
		fmt.Printf("✅ %s é VÁLIDO!  | Formatado: %s\n", numeroInput, num)
		return
	}
	// Tentando analisar o número sem o "+" para detectar o país
	num, err = dto.ValidateCellNumber(fmt.Sprintf("+%s", numeroInput))

	if err == nil {
		fmt.Printf("✅ %s é VÁLIDO!  | Formatado: %s\n", numeroInput, num)
		return
	}

	fmt.Printf("❌ %s: Erro de análise: %v\n", numeroInput, err)

}

func main() {
	// Testando números de vários países sem especificar a região no código
	validarNumeroInternacional("(11)980876112") // Brasil
	validarNumeroInternacional("41 76 752-0704")  // Suiça
	validarNumeroInternacional("+5511999998888") // Brasil
	validarNumeroInternacional("+1813733-9523")   // Estados Unidos
	validarNumeroInternacional("+351211234567")  // Portugal
	validarNumeroInternacional("351912345678")  // Reino Unido
	validarNumeroInternacional("11999998888")    // Vai falhar (sem o +)
}
