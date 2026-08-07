package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// 1. Definição do Endpoint oficial (Homologação ou Produção)
	// Novo WSDL unificado que comporta os layouts atualizados da prefeitura
	url := "https://nfews.prefeitura.sp.gov.br/lotenfe.asmx"

	// 2. Carregar o Certificado Digital A1 (Chave e Certificado em formato .pem)
	cert, err := tls.LoadX509KeyPair("certificado_publico.pem", "chave_privada.pem")
	if err != nil {
		fmt.Printf("Erro ao carregar certificado: %v\n", err)
		return
	}

	// Configurar o cliente HTTP para anexar o Certificado mTLS na requisição
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Prefeitura exige TLS 1.2 ou superior
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	// 4. Executar a requisição HTTP POST para o WebService
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(soapEnvelope))
	if err != nil {
		fmt.Printf("Erro ao criar requisição: %v\n", err)
		return
	}

	// Definir Headers obrigatórios do protocolo SOAP do município
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://prefeitura.sp.gov.br") // Ação mapeada no WSDL

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Erro na comunicação com a Prefeitura: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 5. Ler o XML de retorno enviado pela prefeitura
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler resposta: %v\n", err)
		return
	}

	fmt.Printf("Status HTTP: %s\n", resp.Status)
	fmt.Printf("XML de Resposta:\n%s\n", string(body))
}
