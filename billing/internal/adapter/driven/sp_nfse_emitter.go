package driven

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata"

	"billing/internal/domain"
	"billing/internal/port"

	"github.com/beevik/etree"
	"github.com/russellhaering/goxmldsig"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// SPNfseEmitter handles single RPS submission to São Paulo's municipal web service (Leiaute 2)
type SPNfseEmitter struct {
	logger        port.Logger
	env           string
	certPath      string
	certPassword  string
	serviceCode   int
	aliquot       float64
	serie         string
	timeout       time.Duration
	endpointURL   string
	cLocPrestacao string
	cstIBSCBS     string
	cClassifTrib  string
	cIndOp        string
}

// NewSPNfseEmitter creates a new instance of SPNfseEmitter
func NewSPNfseEmitter(
	logger port.Logger,
	env, certPath, certPassword string,
	serviceCode int,
	aliquot float64,
	serie string,
	timeoutSeconds int,
	endpointURL string,
) *SPNfseEmitter {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 25
	}
	return &SPNfseEmitter{
		logger:        logger,
		env:           env,
		certPath:      certPath,
		certPassword:  certPassword,
		serviceCode:   serviceCode,
		aliquot:       aliquot,
		serie:         serie,
		timeout:       time.Duration(timeoutSeconds) * time.Second,
		endpointURL:   endpointURL,
		cLocPrestacao: "3550308", // Default: São Paulo
		cstIBSCBS:     "499",     // Default: Simples Nacional puro
		cClassifTrib:  "499001",  // Default: Simples Nacional puro
		cIndOp:        "100301",  // Default: serviço consumido no município do prestador
	}
}

// SetCIndOp updates the operation indicator code (cIndOp) of the IBSCBS group
func (e *SPNfseEmitter) SetCIndOp(cIndOp string) {
	if cIndOp != "" {
		e.cIndOp = cIndOp
	}
}

// SetIBSCBS updates the IBS/CBS parameters for Leiaute 2
func (e *SPNfseEmitter) SetIBSCBS(cLocPrestacao, cstIBSCBS, cClassifTrib string) {
	if cLocPrestacao != "" {
		e.cLocPrestacao = cLocPrestacao
	}
	if cstIBSCBS != "" {
		e.cstIBSCBS = cstIBSCBS
	}
	if cClassifTrib != "" {
		e.cClassifTrib = cClassifTrib
	}
}

// EmitSingle sends a single RPS to Prefeitura de SP and returns the emitted NFS-e details
func (e *SPNfseEmitter) EmitSingle(ctx context.Context, invoice *domain.Invoice, rpsNumber int64) (*domain.EmissionItem, error) {
	e.logger.IPrintf(1, "Starting NFS-e emission for Invoice ID %d, RPS #%d (Env: %s)", invoice.ID, rpsNumber, e.env)

	cleanNumbers := regexp.MustCompile(`[^0-9]`)
	if cleanNumbers.ReplaceAllString(invoice.Vendor.Document, "") == "" {
		return nil, fmt.Errorf("CNPJ do prestador (Vendor.Document) não informado na fatura")
	}
	if cleanNumbers.ReplaceAllString(invoice.Vendor.TaxDocument, "") == "" {
		return nil, fmt.Errorf("Inscrição Municipal do prestador (Vendor.TaxDocument/CCM) não informada na fatura")
	}

	if !invoice.IsPF() {
		return nil, fmt.Errorf("emissão de NFS-e não permitida para pessoa jurídica (apenas pessoa física com endereço em branco é suportada no momento)")
	}
	if !invoice.HasNBS() {
		return nil, fmt.Errorf("não é permitido emitir NFS-e: todos os itens da fatura devem possuir código NBS cadastrado")
	}

	tlsCert, rsaKey, x509Cert, err := e.loadCertificates()
	if err != nil {
		e.logger.IPrintf(1, "Failed to load certificate: %v", err)
		return nil, fmt.Errorf("falha ao carregar certificado digital A1: %w", err)
	}

	// 1. Calculate RPS signature (SHA-1 + RSA of fixed-width ASCII string)
	// PMSP requires current local São Paulo date (cannot be ahead of SP date)
	emissionDate := time.Now().In(spLocation())
	rpsSignature, err := e.buildAndSignRPS(invoice, rpsNumber, emissionDate, rsaKey)
	if err != nil {
		e.logger.IPrintf(1, "Failed to sign RPS: %v", err)
		return nil, fmt.Errorf("falha ao assinar RPS: %w", err)
	}

	// 2. Build PedidoEnvioRPS XML
	rawXML, err := e.buildPedidoEnvioRPSXML(invoice, rpsNumber, emissionDate, rpsSignature)
	if err != nil {
		e.logger.IPrintf(1, "Failed to build XML: %v", err)
		return nil, fmt.Errorf("falha ao gerar XML do RPS: %w", err)
	}

	// 3. Sign XML using XMLDSig (Enveloped)
	signedXML, err := e.signXML(rawXML, tlsCert, x509Cert)
	if err != nil {
		e.logger.IPrintf(1, "Failed to sign XML document: %v", err)
		return nil, fmt.Errorf("falha na assinatura digital do XML (XMLDSig): %w", err)
	}

	// 4. Send SOAP request via HTTPS (with mTLS)
	respXML, err := e.sendSOAP(ctx, tlsCert, signedXML)
	if err != nil {
		e.logger.IPrintf(1, "SOAP request to Prefeitura SP failed: %v", err)
		return nil, fmt.Errorf("erro na comunicação com a Prefeitura de SP: %w", err)
	}

	// 5. Parse city response
	item, err := e.parseResponse(respXML, rpsNumber, invoice)
	if err != nil {
		e.logger.IPrintf(1, "Prefeitura SP rejected or returned error: %v", err)
		return nil, err
	}

	e.logger.IPrintf(1, "NFS-e successfully emitted for Invoice ID %d: NFe #%d, Verification: %s",
		invoice.ID, *item.NFENumber, *item.NFEVerification)

	return item, nil
}

func (e *SPNfseEmitter) loadCertificates() (*tls.Certificate, *rsa.PrivateKey, *x509.Certificate, error) {
	password := e.certPassword
	if password == "" {
		password = os.Getenv("BILLING_TAX_CERT_PASSWORD")
	}

	pfxData, err := os.ReadFile(e.certPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("arquivo de certificado não encontrado em %s: %w", e.certPath, err)
	}

	privKey, cert, caCerts, err := pkcs12.DecodeChain(pfxData, password)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("erro ao decodificar arquivo .pfx (verifique a senha): %w", err)
	}

	rsaKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, nil, fmt.Errorf("chave privada do certificado não é RSA")
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  rsaKey,
		Leaf:        cert,
	}
	for _, ca := range caCerts {
		tlsCert.Certificate = append(tlsCert.Certificate, ca.Raw)
	}

	return tlsCert, rsaKey, cert, nil
}

// buildAndSignRPS creates the concatenated ASCII string and signs it with RSA-SHA1 (Leiaute 2)
func (e *SPNfseEmitter) buildAndSignRPS(invoice *domain.Invoice, rpsNumber int64, emissionDate time.Time, rsaKey *rsa.PrivateKey) (string, error) {
	cleanNumbers := regexp.MustCompile(`[^0-9]`)

	ccm := cleanNumbers.ReplaceAllString(invoice.Vendor.TaxDocument, "")
	if len(ccm) > 8 {
		ccm = ccm[:8]
	}
	ccmFormatted := fmt.Sprintf("%08s", ccm)

	serieFormatted := fmt.Sprintf("%-5s", e.serie)
	if len(serieFormatted) > 5 {
		serieFormatted = serieFormatted[:5]
	}

	rpsFormatted := fmt.Sprintf("%012d", rpsNumber)
	dateFormatted := emissionDate.Format("20060102")
	tributacao := "T"
	status := "N"
	issRetido := "N"

	amountCents := int64(invoice.Amount * 100)
	amountFormatted := fmt.Sprintf("%015d", amountCents)
	deductionsFormatted := fmt.Sprintf("%015d", 0)
	serviceCodeFormatted := fmt.Sprintf("%05d", e.serviceCode)

	doc := ""
	if invoice.Customer.Document != nil {
		doc = cleanNumbers.ReplaceAllString(*invoice.Customer.Document, "")
	}
	docType := "1"
	if len(doc) > 11 {
		docType = "2"
	}
	docFormatted := fmt.Sprintf("%014s", doc)

	// Concatenation string defined by PMSP Manual for Layout 2
	asciiStr := ccmFormatted + serieFormatted + rpsFormatted + dateFormatted +
		tributacao + status + issRetido + amountFormatted + deductionsFormatted +
		serviceCodeFormatted + docType + docFormatted

	// Hash SHA-1
	h := sha1.New()
	h.Write([]byte(asciiStr))
	digest := h.Sum(nil)

	// RSA Sign
	sig, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA1, digest)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

func (e *SPNfseEmitter) buildPedidoEnvioRPSXML(invoice *domain.Invoice, rpsNumber int64, emissionDate time.Time, rpsSignature string) (string, error) {
	cleanNumbers := regexp.MustCompile(`[^0-9]`)
	vendorCNPJ := cleanNumbers.ReplaceAllString(invoice.Vendor.Document, "")
	if len(vendorCNPJ) > 0 && len(vendorCNPJ) < 14 {
		vendorCNPJ = fmt.Sprintf("%014s", vendorCNPJ)
	}
	if vendorCNPJ == "" {
		return "", fmt.Errorf("CNPJ do prestador (Vendor.Document) está vazio")
	}

	ccm := cleanNumbers.ReplaceAllString(invoice.Vendor.TaxDocument, "")
	if len(ccm) > 0 && len(ccm) < 8 {
		ccm = fmt.Sprintf("%08s", ccm)
	}
	if len(ccm) > 8 {
		ccm = ccm[:8]
	}
	if ccm == "" {
		return "", fmt.Errorf("Inscrição Municipal do prestador (Vendor.TaxDocument/CCM) está vazia")
	}

	customerDoc := ""
	if invoice.Customer.Document != nil {
		customerDoc = cleanNumbers.ReplaceAllString(*invoice.Customer.Document, "")
	}
	docTag := "CPF"
	if len(customerDoc) > 11 {
		return "", fmt.Errorf("emissão não permitida para tomador PJ (CNPJ %s)", customerDoc)
	}
	if len(customerDoc) > 0 && len(customerDoc) < 11 {
		customerDoc = fmt.Sprintf("%011s", customerDoc)
	}
	if customerDoc == "" {
		return "", fmt.Errorf("documento (CPF) do tomador (Customer.Document) está vazio")
	}

	customerEmail := ""
	if invoice.Customer.Email != nil {
		customerEmail = strings.TrimSpace(*invoice.Customer.Email)
	}

	// Extract primary NBS code from invoice items
	var nbsCode int64
	for _, item := range invoice.InvoiceItems {
		if item.NBSCode != nil && *item.NBSCode > 0 {
			nbsCode = *item.NBSCode
			break
		}
	}
	if nbsCode <= 0 {
		return "", fmt.Errorf("nenhum item da fatura possui código NBS informado")
	}
	nbsFormatted := fmt.Sprintf("%09d", nbsCode)

	cLoc := e.cLocPrestacao
	if cLoc == "" {
		cLoc = "3550308"
	}
	// Note: PMSP Layout 2 (tpGIBSCBS) has no CST element; the CST is implied by cClassTrib.
	cClassif := e.cClassifTrib
	if cClassif == "" {
		cClassif = "499001"
	}
	cIndOp := e.cIndOp
	if cIndOp == "" {
		cIndOp = "100301"
	}

	// Build description
	var descBuilder strings.Builder
	for _, item := range invoice.InvoiceItems {
		descBuilder.WriteString(fmt.Sprintf("%dx %s | ", item.Quantity, item.Description))
	}
	desc := strings.TrimSuffix(descBuilder.String(), " | ")
	if desc == "" {
		desc = "Prestação de serviços"
	}

	// XML Escaping
	var bufDesc bytes.Buffer
	_ = xml.EscapeText(&bufDesc, []byte(desc))
	escapedDesc := bufDesc.String()

	var bufName bytes.Buffer
	_ = xml.EscapeText(&bufName, []byte(invoice.Customer.Name))
	escapedName := bufName.String()

	// PedidoEnvioRPS layout 2 PMSP
	xmlContent := fmt.Sprintf(`<PedidoEnvioRPS xmlns="http://www.prefeitura.sp.gov.br/nfe">
  <Cabecalho xmlns="" Versao="2">
    <CPFCNPJRemetente>
      <CNPJ>%s</CNPJ>
    </CPFCNPJRemetente>
  </Cabecalho>
  <RPS xmlns="">
    <Assinatura>%s</Assinatura>
    <ChaveRPS>
      <InscricaoPrestador>%s</InscricaoPrestador>
      <SerieRPS>%s</SerieRPS>
      <NumeroRPS>%d</NumeroRPS>
    </ChaveRPS>
    <TipoRPS>RPS</TipoRPS>
    <DataEmissao>%s</DataEmissao>
    <StatusRPS>N</StatusRPS>
    <TributacaoRPS>T</TributacaoRPS>
    <ValorDeducoes>0.00</ValorDeducoes>
    <ValorPIS>0.00</ValorPIS>
    <ValorCOFINS>0.00</ValorCOFINS>
    <ValorINSS>0.00</ValorINSS>
    <ValorIR>0.00</ValorIR>
    <ValorCSLL>0.00</ValorCSLL>
    <CodigoServico>%05d</CodigoServico>
    <AliquotaServicos>%.4f</AliquotaServicos>
    <ISSRetido>false</ISSRetido>
    <CPFCNPJTomador>
      <%s>%s</%s>
    </CPFCNPJTomador>
    <RazaoSocialTomador>%s</RazaoSocialTomador>
    <EmailTomador>%s</EmailTomador>
    <Discriminacao>%s</Discriminacao>
    <ValorFinalCobrado>%.2f</ValorFinalCobrado>
    <ValorIPI>0.00</ValorIPI>
    <ExigibilidadeSuspensa>0</ExigibilidadeSuspensa>
    <PagamentoParceladoAntecipado>0</PagamentoParceladoAntecipado>
    <NBS>%s</NBS>
    <cLocPrestacao>%s</cLocPrestacao>
    <IBSCBS>
      <finNFSe>0</finNFSe>
      <indFinal>1</indFinal>
      <cIndOp>%s</cIndOp>
      <indDest>0</indDest>
      <valores>
        <trib>
          <gIBSCBS>
            <cClassTrib>%s</cClassTrib>
          </gIBSCBS>
        </trib>
      </valores>
    </IBSCBS>
  </RPS>
</PedidoEnvioRPS>`,
		vendorCNPJ,
		rpsSignature,
		ccm,
		e.serie,
		rpsNumber,
		emissionDate.Format("2006-01-02"),
		e.serviceCode,
		e.aliquot,
		docTag, customerDoc, docTag,
		escapedName,
		customerEmail,
		escapedDesc,
		invoice.Amount,
		nbsFormatted,
		cLoc,
		cIndOp,
		cClassif,
	)

	return xmlContent, nil
}

func (e *SPNfseEmitter) signXML(xmlContent string, tlsCert *tls.Certificate, _ *x509.Certificate) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return "", err
	}

	root := doc.Root()
	if root == nil {
		return "", fmt.Errorf("XML vazio ou inválido")
	}

	signingContext := dsig.NewDefaultSigningContext(dsig.TLSCertKeyStore(*tlsCert))
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	signingContext.Hash = crypto.SHA1
	signingContext.Prefix = ""

	signedRoot, err := signingContext.SignEnveloped(root)
	if err != nil {
		return "", fmt.Errorf("erro ao assinar envelope XML: %w", err)
	}

	signedDoc := etree.NewDocument()
	signedDoc.SetRoot(signedRoot)
	signedDoc.WriteSettings.UseCRLF = false

	return signedDoc.WriteToString()
}

func (e *SPNfseEmitter) sendSOAP(ctx context.Context, clientCert *tls.Certificate, xmlContent string) (string, error) {
	soapEnvelope := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <EnvioRPSRequest xmlns="http://www.prefeitura.sp.gov.br/nfe">
      <VersaoSchema>2</VersaoSchema>
      <MensagemXML><![CDATA[%s]]></MensagemXML>
    </EnvioRPSRequest>
  </soap:Body>
</soap:Envelope>`, xmlContent)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:  []tls.Certificate{*clientCert},
			MinVersion:    tls.VersionTLS12,
			Renegotiation: tls.RenegotiateFreelyAsClient,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   e.timeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpointURL, bytes.NewBufferString(soapEnvelope))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://www.prefeitura.sp.gov.br/nfe/ws/envioRPS")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return string(bodyBytes), nil
}

type soapEnvioRPSResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    soapBody `xml:"Body"`
}

type soapBody struct {
	EnvioRPSResponse envioRPSResponse `xml:"EnvioRPSResponse"`
}

type envioRPSResponse struct {
	RetornoXML string `xml:"RetornoXML"`
}

type retornoEnvioRPS struct {
	XMLName   xml.Name `xml:"RetornoEnvioRPS"`
	Cabecalho struct {
		Sucesso bool `xml:"Sucesso"`
	} `xml:"Cabecalho"`
	Erros []struct {
		Codigo    int    `xml:"Codigo"`
		Descricao string `xml:"Descricao"`
	} `xml:"Erro"`
	Alertas []struct {
		Codigo    int    `xml:"Codigo"`
		Descricao string `xml:"Descricao"`
	} `xml:"Alerta"`
	ChaveNFeRPS []struct {
		ChaveNFe struct {
			NumeroNFe         int64  `xml:"NumeroNFe"`
			CodigoVerificacao string `xml:"CodigoVerificacao"`
		} `xml:"ChaveNFe"`
		ChaveRPS struct {
			NumeroRPS int64 `xml:"NumeroRPS"`
		} `xml:"ChaveRPS"`
	} `xml:"ChaveNFeRPS"`
}

func (e *SPNfseEmitter) parseResponse(soapResponse string, rpsNumber int64, invoice *domain.Invoice) (*domain.EmissionItem, error) {
	var soapResp soapEnvioRPSResponse
	if err := xml.Unmarshal([]byte(soapResponse), &soapResp); err != nil {
		trimmed := strings.TrimSpace(soapResponse)
		if strings.HasPrefix(trimmed, "<!DOCTYPE html") || strings.HasPrefix(trimmed, "<html") || strings.Contains(trimmed, "<html") {
			return nil, fmt.Errorf("o servidor retornou uma página HTML em vez de resposta SOAP XML (verifique a URL do WebService): %w", err)
		}
		return nil, fmt.Errorf("resposta SOAP inválida: %w", err)
	}

	retornoXMLContent := strings.TrimSpace(soapResp.Body.EnvioRPSResponse.RetornoXML)
	if retornoXMLContent == "" {
		// Could be raw RetornoEnvioRPS
		retornoXMLContent = soapResponse
	}

	var retorno retornoEnvioRPS
	if err := xml.Unmarshal([]byte(retornoXMLContent), &retorno); err != nil {
		return nil, fmt.Errorf("falha ao interpretar RetornoXML da Prefeitura: %w (XML: %s)", err, retornoXMLContent)
	}

	if !retorno.Cabecalho.Sucesso {
		var errMsgs []string
		for _, erro := range retorno.Erros {
			errMsgs = append(errMsgs, fmt.Sprintf("[%d] %s", erro.Codigo, strings.TrimSpace(erro.Descricao)))
		}
		if len(errMsgs) == 0 {
			errMsgs = append(errMsgs, "Erro desconhecido retornado pela Prefeitura")
		}
		return nil, fmt.Errorf("rejeição pela Prefeitura de SP: %s", strings.Join(errMsgs, "; "))
	}

	if len(retorno.ChaveNFeRPS) == 0 {
		return nil, fmt.Errorf("prefeitura confirmou sucesso, mas não retornou os dados da NFS-e")
	}

	chave := retorno.ChaveNFeRPS[0]
	nfeNumber := chave.ChaveNFe.NumeroNFe
	nfeVerif := chave.ChaveNFe.CodigoVerificacao
	now := time.Now().In(spLocation())
	amount := invoice.Amount

	return &domain.EmissionItem{
		InvoiceID:       invoice.ID,
		RPSNumber:       rpsNumber,
		NFENumber:       &nfeNumber,
		NFEDatetime:     &now,
		NFEVerification: &nfeVerif,
		NFEAmount:       &amount,
	}, nil
}

// spLocation returns the America/Sao_Paulo time.Location (-03:00)
func spLocation() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err == nil {
		return loc
	}
	return time.FixedZone("America/Sao_Paulo", -3*3600)
}
