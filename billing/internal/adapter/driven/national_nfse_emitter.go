package driven

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"billing/internal/domain"
	"billing/internal/port"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

const (
	nationalNFSeNamespace          = "http://www.sped.fazenda.gov.br/nfse"
	defaultNationalProductionURL   = "https://sefin.nfse.gov.br/SefinNacional"
	defaultNationalHomologationURL = "https://sefin.producaorestrita.nfse.gov.br/SefinNacional"
	defaultNationalDPSVersion      = "1.00"
	defaultNationalSerie           = "1"
	nationalAppVersion             = "billing-1.0"
)

var onlyDigits = regexp.MustCompile(`[^0-9]`)

// NationalNfseConfig holds the parameters of the NFS-e Nacional (Sefin Nacional) emitter.
type NationalNfseConfig struct {
	Env            string // "production" => tpAmb 1; anything else => tpAmb 2 (produção restrita)
	CertPath       string
	CertPassword   string
	BaseURL        string // e.g. https://sefin.nfse.gov.br/SefinNacional (the emitter posts to BaseURL + "/nfse")
	TimeoutSeconds int
	DPSVersion     string // DPS layout version attribute (default 1.00)
	Serie          string // numeric DPS series, 1..99999 (default 1)
	CLocEmi        string // IBGE code of the issuer municipality (default 3550308 - São Paulo)
	CLocPrestacao  string // IBGE code of the place of service (default = CLocEmi)
	CTribNac       string // 6-digit national taxation code (e.g. 080201)
	CTribMun       string // 3-digit municipal breakdown of cTribNac (e.g. 003 for 08.02.01.003 in São Paulo); optional
	OpSimpNac      int    // 1 = não optante, 2 = MEI, 3 = ME/EPP (default 3)
	RegApTribSN    int    // apuração for ME/EPP: 1 = tributos federais e municipal pelo SN (default 1)
	RegEspTrib     int    // regime especial de tributação (default 0 = nenhum)
	IncludeIM      bool   // send the vendor's municipal registration (prest/IM)
}

// NationalNfseEmitter issues a single NFS-e through the Sefin Nacional API (DPS -> NFS-e).
// It is required for ME/EPP optantes pelo Simples Nacional in São Paulo from 01/11/2026 (IN SF/SUREM 13/2026).
type NationalNfseEmitter struct {
	logger    port.Logger
	cfg       NationalNfseConfig
	now       func() time.Time
	transport *http.Transport // optional override (tests)
}

// NewNationalNfseEmitter creates a new NationalNfseEmitter applying defaults to empty settings.
func NewNationalNfseEmitter(logger port.Logger, cfg NationalNfseConfig) *NationalNfseEmitter {
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 30
	}
	if cfg.BaseURL == "" {
		if cfg.Env == "production" {
			cfg.BaseURL = defaultNationalProductionURL
		} else {
			cfg.BaseURL = defaultNationalHomologationURL
		}
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.DPSVersion == "" {
		cfg.DPSVersion = defaultNationalDPSVersion
	}
	cfg.Serie = strings.TrimLeft(onlyDigits.ReplaceAllString(cfg.Serie, ""), "0")
	if cfg.Serie == "" {
		cfg.Serie = defaultNationalSerie
	}
	if cfg.CLocEmi == "" {
		cfg.CLocEmi = "3550308"
	}
	if cfg.CLocPrestacao == "" {
		cfg.CLocPrestacao = cfg.CLocEmi
	}
	if cfg.OpSimpNac == 0 {
		cfg.OpSimpNac = 3
	}
	if cfg.OpSimpNac == 3 && cfg.RegApTribSN == 0 {
		cfg.RegApTribSN = 1
	}
	return &NationalNfseEmitter{logger: logger, cfg: cfg, now: time.Now}
}

func (e *NationalNfseEmitter) tpAmb() int {
	if e.cfg.Env == "production" {
		return 1
	}
	return 2
}

// EmitSingle builds, signs and sends a DPS to the Sefin Nacional and returns the emitted NFS-e.
// The DPS number (nDPS) is the rpsNumber; NFEVerification stores the 50-digit chave de acesso.
func (e *NationalNfseEmitter) EmitSingle(ctx context.Context, invoice *domain.Invoice, rpsNumber int64) (*domain.EmissionItem, error) {
	e.logger.IPrintf(1, "Starting NFS-e Nacional emission for Invoice ID %d, DPS #%d (série %s, tpAmb %d)",
		invoice.ID, rpsNumber, e.cfg.Serie, e.tpAmb())

	if !invoice.IsPF() {
		return nil, fmt.Errorf("emissão de NFS-e não permitida para pessoa jurídica (apenas pessoa física é suportada no momento)")
	}
	if !invoice.HasNBS() {
		return nil, fmt.Errorf("não é permitido emitir NFS-e: todos os itens da fatura devem possuir código NBS cadastrado")
	}

	tlsCert, _, _, err := loadA1Certificate(e.cfg.CertPath, e.cfg.CertPassword)
	if err != nil {
		e.logger.IPrintf(1, "Failed to load certificate: %v", err)
		return nil, fmt.Errorf("falha ao carregar certificado digital A1: %w", err)
	}

	dpsXML, dpsID, err := e.buildDPSXML(invoice, rpsNumber, e.now().In(spLocation()))
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar XML da DPS: %w", err)
	}

	signedXML, err := signDPS(dpsXML, tlsCert)
	if err != nil {
		return nil, fmt.Errorf("falha na assinatura digital da DPS (XMLDSig): %w", err)
	}

	resp, err := e.post(ctx, tlsCert, signedXML)
	if err != nil {
		e.logger.IPrintf(1, "Sefin Nacional request failed for Invoice ID %d (%s): %v", invoice.ID, dpsID, err)
		return nil, err
	}

	item, err := e.parseResult(resp, rpsNumber, invoice)
	if err != nil {
		return nil, err
	}
	e.logger.IPrintf(1, "NFS-e Nacional emitted for Invoice ID %d: nNFSe %d, chave %s", invoice.ID, *item.NFENumber, *item.NFEVerification)
	return item, nil
}

// buildDPSXML returns the unsigned DPS document and its infDPS Id.
func (e *NationalNfseEmitter) buildDPSXML(invoice *domain.Invoice, nDPS int64, now time.Time) (string, string, error) {
	if nDPS <= 0 {
		return "", "", fmt.Errorf("número da DPS inválido: %d", nDPS)
	}
	vendorCNPJ := onlyDigits.ReplaceAllString(invoice.Vendor.Document, "")
	if len(vendorCNPJ) != 14 {
		return "", "", fmt.Errorf("CNPJ do prestador (Vendor.Document) inválido: %q", invoice.Vendor.Document)
	}
	cTribNac := onlyDigits.ReplaceAllString(e.cfg.CTribNac, "")
	if len(cTribNac) != 6 {
		return "", "", fmt.Errorf("código de tributação nacional (cTribNac) deve ter 6 dígitos; configurado: %q", e.cfg.CTribNac)
	}
	cTribMun := onlyDigits.ReplaceAllString(e.cfg.CTribMun, "")
	if cTribMun != "" && len(cTribMun) != 3 {
		return "", "", fmt.Errorf("código de tributação municipal (cTribMun) deve ter 3 dígitos; configurado: %q", e.cfg.CTribMun)
	}
	customerDoc := ""
	if invoice.Customer.Document != nil {
		customerDoc = onlyDigits.ReplaceAllString(*invoice.Customer.Document, "")
	}
	if len(customerDoc) > 11 {
		return "", "", fmt.Errorf("emissão não permitida para tomador PJ (CNPJ %s)", customerDoc)
	}
	if customerDoc == "" {
		return "", "", fmt.Errorf("documento (CPF) do tomador (Customer.Document) está vazio")
	}
	customerDoc = fmt.Sprintf("%011s", customerDoc)

	var nbs int64
	for _, item := range invoice.InvoiceItems {
		if item.NBSCode != nil && *item.NBSCode > 0 {
			nbs = *item.NBSCode
			break
		}
	}

	// Id = "DPS" + cLocEmi(7) + tpInsc(1: 2=CNPJ) + inscFed(14) + série(5) + nDPS(15)
	serieNum, _ := strconv.Atoi(e.cfg.Serie)
	dpsID := fmt.Sprintf("DPS%s2%s%05d%015d", e.cfg.CLocEmi, vendorCNPJ, serieNum, nDPS)

	// dhEmi must not be ahead of the Sefin server clock
	dhEmi := now.Add(-1 * time.Minute)

	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	root := doc.CreateElement("DPS")
	root.CreateAttr("xmlns", nationalNFSeNamespace)
	root.CreateAttr("versao", e.cfg.DPSVersion)

	inf := root.CreateElement("infDPS")
	inf.CreateAttr("Id", dpsID)
	add := func(parent *etree.Element, tag, text string) *etree.Element {
		el := parent.CreateElement(tag)
		if text != "" {
			el.SetText(text)
		}
		return el
	}

	add(inf, "tpAmb", strconv.Itoa(e.tpAmb()))
	add(inf, "dhEmi", dhEmi.Format("2006-01-02T15:04:05-07:00"))
	add(inf, "verAplic", nationalAppVersion)
	add(inf, "serie", e.cfg.Serie)
	add(inf, "nDPS", strconv.FormatInt(nDPS, 10))
	add(inf, "dCompet", now.Format("2006-01-02"))
	add(inf, "tpEmit", "1")
	add(inf, "cLocEmi", e.cfg.CLocEmi)

	prest := add(inf, "prest", "")
	add(prest, "CNPJ", vendorCNPJ)
	if e.cfg.IncludeIM {
		if im := onlyDigits.ReplaceAllString(invoice.Vendor.TaxDocument, ""); im != "" {
			add(prest, "IM", im)
		}
	}
	regTrib := add(prest, "regTrib", "")
	add(regTrib, "opSimpNac", strconv.Itoa(e.cfg.OpSimpNac))
	if e.cfg.OpSimpNac == 3 {
		add(regTrib, "regApTribSN", strconv.Itoa(e.cfg.RegApTribSN))
	}
	add(regTrib, "regEspTrib", strconv.Itoa(e.cfg.RegEspTrib))

	toma := add(inf, "toma", "")
	add(toma, "CPF", customerDoc)
	add(toma, "xNome", truncateRunes(strings.TrimSpace(invoice.Customer.Name), 300))
	if invoice.Customer.Email != nil {
		if email := strings.TrimSpace(*invoice.Customer.Email); email != "" {
			add(toma, "email", truncateRunes(email, 80))
		}
	}

	serv := add(inf, "serv", "")
	locPrest := add(serv, "locPrest", "")
	add(locPrest, "cLocPrestacao", e.cfg.CLocPrestacao)
	cServ := add(serv, "cServ", "")
	add(cServ, "cTribNac", cTribNac)
	if cTribMun != "" {
		add(cServ, "cTribMun", cTribMun)
	}
	add(cServ, "xDescServ", truncateRunes(buildServiceDescription(invoice), 2000))
	if nbs > 0 {
		add(cServ, "cNBS", fmt.Sprintf("%09d", nbs))
	}

	valores := add(inf, "valores", "")
	vServPrest := add(valores, "vServPrest", "")
	add(vServPrest, "vServ", fmt.Sprintf("%.2f", invoice.Amount))
	trib := add(valores, "trib", "")
	tribMun := add(trib, "tribMun", "")
	add(tribMun, "tribISSQN", "1")  // 1 = operação tributável
	add(tribMun, "tpRetISSQN", "1") // 1 = não retido
	totTrib := add(trib, "totTrib", "")
	add(totTrib, "indTotTrib", "0") // não informar valor estimado dos tributos (Decreto 8.264/2014)

	doc.WriteSettings.CanonicalText = true
	out, err := doc.WriteToString()
	if err != nil {
		return "", "", err
	}
	return out, dpsID, nil
}

func buildServiceDescription(invoice *domain.Invoice) string {
	parts := make([]string, 0, len(invoice.InvoiceItems))
	for _, item := range invoice.InvoiceItems {
		parts = append(parts, fmt.Sprintf("%dx %s", item.Quantity, strings.TrimSpace(item.Description)))
	}
	desc := strings.Join(parts, " | ")
	if desc == "" {
		desc = "Prestação de serviços"
	}
	return desc
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}

// signDPS signs infDPS (Reference URI="#Id", enveloped + C14N 1.0) and appends the
// Signature as the last child of DPS, as required by the NFS-e Nacional XSD.
func signDPS(dpsXML string, tlsCert *tls.Certificate) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(dpsXML); err != nil {
		return "", err
	}
	root := doc.Root()
	if root == nil {
		return "", fmt.Errorf("XML da DPS vazio")
	}
	inf := root.SelectElement("infDPS")
	if inf == nil {
		return "", fmt.Errorf("elemento infDPS não encontrado")
	}

	ctx := dsig.NewDefaultSigningContext(dsig.TLSCertKeyStore(*tlsCert))
	ctx.Canonicalizer = dsig.MakeC14N10RecCanonicalizer()
	ctx.Hash = crypto.SHA1
	ctx.Prefix = ""
	ctx.IdAttribute = "Id"

	sig, err := ctx.ConstructSignature(inf, true)
	if err != nil {
		return "", err
	}
	root.AddChild(sig)

	doc.WriteSettings.CanonicalText = true
	return doc.WriteToString()
}

type nationalMessage struct {
	Codigo      string `json:"codigo"`
	Descricao   string `json:"descricao"`
	Complemento string `json:"complemento"`
}

type nationalResponse struct {
	ChaveAcesso    string            `json:"chaveAcesso"`
	IDDPS          string            `json:"idDps"`
	NFSeXMLGZipB64 string            `json:"nfseXmlGZipB64"`
	Alertas        []nationalMessage `json:"alertas"`
	Erros          []nationalMessage `json:"erros"`
	Erro           *nationalMessage  `json:"erro"`
}

func (m nationalMessage) String() string {
	s := fmt.Sprintf("[%s] %s", m.Codigo, strings.TrimSpace(m.Descricao))
	if c := strings.TrimSpace(m.Complemento); c != "" {
		s += " (" + c + ")"
	}
	return s
}

func (e *NationalNfseEmitter) post(ctx context.Context, tlsCert *tls.Certificate, signedXML string) (*nationalResponse, error) {
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	if _, err := zw.Write([]byte(signedXML)); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	body, _ := json.Marshal(map[string]string{"dpsXmlGZipB64": base64.StdEncoding.EncodeToString(gz.Bytes())})

	transport := e.transport
	if transport == nil {
		transport = &http.Transport{TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{*tlsCert},
			MinVersion:   tls.VersionTLS12,
		}}
	}
	client := &http.Client{
		Timeout:   time.Duration(e.cfg.TimeoutSeconds) * time.Second,
		Transport: transport,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.cfg.BaseURL+"/nfse", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na comunicação com a Sefin Nacional (NFS-e Nacional): %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta da Sefin Nacional: %w", err)
	}

	var out nationalResponse
	jsonErr := json.Unmarshal(raw, &out)
	if resp.StatusCode < 200 || resp.StatusCode > 299 || len(out.Erros) > 0 || out.Erro != nil {
		msgs := make([]string, 0, len(out.Erros)+1)
		for _, m := range out.Erros {
			msgs = append(msgs, m.String())
		}
		if out.Erro != nil {
			msgs = append(msgs, out.Erro.String())
		}
		if len(msgs) == 0 {
			msgs = append(msgs, truncateRunes(strings.TrimSpace(string(raw)), 1000))
		}
		return nil, fmt.Errorf("rejeição pela NFS-e Nacional (HTTP %d): %s", resp.StatusCode, strings.Join(msgs, "; "))
	}
	if jsonErr != nil {
		return nil, fmt.Errorf("resposta inválida da Sefin Nacional: %w (%s)", jsonErr, truncateRunes(string(raw), 1000))
	}
	return &out, nil
}

// nfseNumber extracts infNFSe/nNFSe from the gzipped+base64 NFS-e XML returned by the Sefin.
func nfseNumber(nfseXMLGZipB64 string) (int64, error) {
	gz, err := base64.StdEncoding.DecodeString(nfseXMLGZipB64)
	if err != nil {
		return 0, err
	}
	zr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return 0, err
	}
	defer zr.Close()
	var nfse struct {
		Inf struct {
			NNFSe string `xml:"nNFSe"`
		} `xml:"infNFSe"`
	}
	if err := xml.NewDecoder(zr).Decode(&nfse); err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(nfse.Inf.NNFSe), 10, 64)
}

func (e *NationalNfseEmitter) parseResult(resp *nationalResponse, rpsNumber int64, invoice *domain.Invoice) (*domain.EmissionItem, error) {
	if resp.ChaveAcesso == "" {
		return nil, fmt.Errorf("Sefin Nacional não retornou a chave de acesso da NFS-e")
	}
	for _, a := range resp.Alertas {
		e.logger.IPrintf(1, "NFS-e Nacional alert for Invoice ID %d: %s", invoice.ID, a.String())
	}
	number, err := nfseNumber(resp.NFSeXMLGZipB64)
	if err != nil {
		// The note was issued; do not fail. Keep the access key as the reference.
		e.logger.IPrintf(1, "Could not read nNFSe from returned XML (chave %s): %v", resp.ChaveAcesso, err)
		number = 0
	}
	chave := resp.ChaveAcesso
	now := e.now().In(spLocation())
	amount := invoice.Amount
	return &domain.EmissionItem{
		InvoiceID:       invoice.ID,
		RPSNumber:       rpsNumber,
		NFENumber:       &number,
		NFEDatetime:     &now,
		NFEVerification: &chave,
		NFEAmount:       &amount,
	}, nil
}
