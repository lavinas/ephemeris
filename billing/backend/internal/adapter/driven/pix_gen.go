package driven

import (
	"billing/internal/dto"
	"billing/internal/port"
	"fmt"
	"strings"
	"unicode"
	"errors"

	"encoding/base64"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/skip2/go-qrcode"
)

// PixToken is a struct that implements the port.PixToken interface.
type PixToken struct {
	logger port.Logger
}

// NewPixToken creates a new instance of PixToken.
func NewPixToken(logger port.Logger) *PixToken {
	return &PixToken{
		logger: logger,
	}
}

// Get generates the Pix payment payload based on the provided request data.
func (p *PixToken) Get(request port.InDTO) (string, string, error) {
	p.logger.IPrintf(2, "Generating Pix payload for request: %v", request)
	in, ok := request.(*dto.PixRequest)
	if !ok {
		return "", "", errors.New("invalid input type")
	}
	key := p.removeAccents(in.Key)
	desc := p.removeAccents(in.Description)
	name := p.removeAccents(in.Name)
	city := p.removeAccents(in.City)
	Txid := p.removeAccents(in.Txid)

	payload := p.gerarPayloadPix(key, desc, name, city, in.Amount, Txid)
	qrCode, err := p.gerarQRCode(payload)
	if err != nil {
		return "", "", err
	}
	return payload, qrCode, nil
}

// Calcula o CRC-16 (polinômio 0x1021) conforme padrão do Banco Central
func (p *PixToken) calcCRC16(payload string) string {
	crc := 0xFFFF
	polynomial := 0x1021

	for i := 0; i < len(payload); i++ {
		crc ^= int(payload[i]) << 8
		for j := 0; j < 8; j++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ polynomial
			} else {
				crc <<= 1
			}
		}
	}
	return fmt.Sprintf("%04X", crc&0xFFFF)
}

// Formata tag e valor no padrão TLV (Tag, Length, Value)
func (p *PixToken) montaCampo(tag string, valor string) string {
	return fmt.Sprintf("%s%02d%s", tag, len(valor), valor)
}

// Gera o Payload do Pix Copia e Cola
func (p *PixToken) gerarPayloadPix(chave string, descricao string, nome string, cidade string, valor float64, txid string) string {
	// Trata os dados para o padrão aceito pelo Santander
	nomeTratado := strings.ToUpper(p.removeAccents(nome))
	if len(nomeTratado) > 25 {
		nomeTratado = nomeTratado[:25] // Limite padrão EMV
	}

	cidadeTratada := strings.ToUpper(p.removeAccents(cidade))
	if len(cidadeTratada) > 15 {
		cidadeTratada = cidadeTratada[:15] // Limite estrito do BR Code
	}

	if txid == "" {
		txid = "***"
	}

	// Montagem do payload
	payload := p.montaCampo("00", "01")
	payload += p.montaCampo("01", "11")

	merchantAccount := p.montaCampo("00", "BR.GOV.BCB.PIX")
	merchantAccount += p.montaCampo("01", chave)
	if descricao != "" {
		merchantAccount += p.montaCampo("02", descricao)
	}
	payload += p.montaCampo("26", merchantAccount)

	payload += p.montaCampo("52", "0000")
	payload += p.montaCampo("53", "986")

	if valor > 0 {
		payload += p.montaCampo("54", fmt.Sprintf("%.2f", valor))
	}

	payload += p.montaCampo("58", "BR")
	payload += p.montaCampo("59", nomeTratado)
	payload += p.montaCampo("60", cidadeTratada)

	adicional := p.montaCampo("05", txid)
	payload += p.montaCampo("62", adicional)

	payload += "6304"
	payload += p.calcCRC16(payload)

	return payload
}

// removeAccents is a helper function to remove accents from a string.
func (i *PixToken) removeAccents(texto string) string {
	// Transforma a string para separar letras dos acentos, remove os acentos e recompõe
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	// Aplica a transformação
	resultado, _, _ := transform.String(t, texto)

	return resultado
}

// gerarQRCode generates a QR code image from the provided Pix payload and saves it to the specified file path.
func (p *PixToken) gerarQRCode(payload string) (string, error) {
	// Gera os bytes do PNG diretamente na memória
	var pngBytes []byte
	pngBytes, err := qrcode.Encode(payload, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	// Converte os bytes para string Base64
	stringBase64 := base64.StdEncoding.EncodeToString(pngBytes)

	// Retorna no formato pronto para usar na tag <img src="..."> do HTML
	return "data:image/png;base64," + stringBase64, nil
}
