package driven

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"billing/internal/domain"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

func testNationalInvoice() *domain.Invoice {
	doc := "4234322385"
	email := "cliente@teste.com"
	nbs := int64(122051900)
	return &domain.Invoice{
		ID:     7,
		Amount: 150.5,
		Vendor: domain.Vendor{Document: "27.928.875/0001-04", TaxDocument: "5727888-1"},
		Customer: domain.Customer{
			Name:     "Maria & Filhos <Teste>",
			Document: &doc,
			Email:    &email,
		},
		InvoiceItems: []domain.InvoiceItem{{Quantity: 1, Description: "Aula de canto", NBSCode: &nbs}},
	}
}

func testTLSCert(t *testing.T) *tls.Certificate {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "TESTE:27928875000104"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, _ := x509.ParseCertificate(der)
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: cert}
}

func testNationalEmitter(baseURL string) *NationalNfseEmitter {
	e := NewNationalNfseEmitter(&dummyLogger{}, NationalNfseConfig{BaseURL: baseURL, CTribNac: "080201", CTribMun: "003"})
	e.now = func() time.Time { return time.Date(2026, 11, 13, 10, 0, 0, 0, spLocation()) }
	return e
}

func TestNationalNfse_BuildDPSXML(t *testing.T) {
	e := testNationalEmitter("")
	xmlStr, id, err := e.buildDPSXML(testNationalInvoice(), 42, e.now())
	if err != nil {
		t.Fatal(err)
	}
	wantID := "DPS3550308227928875000104" + "00001" + "000000000000042"
	if id != wantID || len(id) != 45 {
		t.Fatalf("id = %s (len %d), want %s", id, len(id), wantID)
	}
	for _, s := range []string{
		`<DPS xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.00">`,
		`<infDPS Id="` + wantID + `">`,
		`<tpAmb>2</tpAmb>`,
		`<dhEmi>2026-11-13T09:59:00-03:00</dhEmi>`,
		`<serie>1</serie>`, `<nDPS>42</nDPS>`, `<dCompet>2026-11-13</dCompet>`,
		`<cLocEmi>3550308</cLocEmi>`,
		`<prest><CNPJ>27928875000104</CNPJ><regTrib><opSimpNac>3</opSimpNac><regApTribSN>1</regApTribSN><regEspTrib>0</regEspTrib></regTrib></prest>`,
		`<CPF>04234322385</CPF>`,
		`<xNome>Maria &amp; Filhos &lt;Teste&gt;</xNome>`,
		`<cTribNac>080201</cTribNac><cTribMun>003</cTribMun><xDescServ>1x Aula de canto</xDescServ><cNBS>122051900</cNBS>`,
		`<vServ>150.50</vServ>`,
		`<tribMun><tribISSQN>1</tribISSQN><tpRetISSQN>1</tpRetISSQN></tribMun>`,
	} {
		if !strings.Contains(xmlStr, s) {
			t.Errorf("DPS XML missing %s\n%s", s, xmlStr)
		}
	}
	if strings.Contains(xmlStr, "<IM>") {
		t.Errorf("IM must be omitted unless IncludeIM is set")
	}

	bad := testNationalEmitter("")
	bad.cfg.CTribNac = "0802"
	if _, _, err := bad.buildDPSXML(testNationalInvoice(), 1, bad.now()); err == nil {
		t.Errorf("expected error for invalid cTribNac")
	}
}

func TestNationalNfse_SignDPSValidates(t *testing.T) {
	e := testNationalEmitter("")
	cert := testTLSCert(t)
	xmlStr, id, err := e.buildDPSXML(testNationalInvoice(), 1, e.now())
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signDPS(xmlStr, cert)
	if err != nil {
		t.Fatal(err)
	}
	if p := os.Getenv("DPS_DUMP"); p != "" {
		_ = os.WriteFile(p, []byte(signed), 0o600)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(signed); err != nil {
		t.Fatal(err)
	}
	root := doc.Root()
	last := root.ChildElements()[len(root.ChildElements())-1]
	if last.Tag != "Signature" {
		t.Fatalf("Signature must be the last child of DPS, got %s", last.Tag)
	}
	if !strings.Contains(signed, `URI="#`+id+`"`) {
		t.Errorf("reference URI must point to infDPS Id")
	}

	// Independent check (the PMSP/Sefin layout puts Signature beside infDPS, which goxmldsig's
	// validator does not support): digest of C14N(infDPS) and RSA-SHA1 over C14N(SignedInfo).
	// Cross-checked against lxml + openssl during development.
	c14n := dsig.MakeC14N10RecCanonicalizer()
	infC14N, err := c14n.Canonicalize(root.SelectElement("infDPS"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha1.Sum(infC14N)
	if got := last.FindElement("./SignedInfo/Reference/DigestValue").Text(); got != base64.StdEncoding.EncodeToString(sum[:]) {
		t.Fatalf("digest mismatch")
	}
	siC14N, err := c14n.Canonicalize(last.SelectElement("SignedInfo"))
	if err != nil {
		t.Fatal(err)
	}
	sigBytes, _ := base64.StdEncoding.DecodeString(last.SelectElement("SignatureValue").Text())
	siSum := sha1.Sum(siC14N)
	if err := rsa.VerifyPKCS1v15(cert.Leaf.PublicKey.(*rsa.PublicKey), crypto.SHA1, siSum[:], sigBytes); err != nil {
		t.Fatalf("signature does not validate: %v", err)
	}
}

func TestNationalNfse_EmitFlow(t *testing.T) {
	cert := testTLSCert(t)

	var nfseGz bytes.Buffer
	zw := gzip.NewWriter(&nfseGz)
	_, _ = zw.Write([]byte(`<?xml version="1.0"?><NFSe xmlns="http://www.sped.fazenda.gov.br/nfse"><infNFSe Id="NFS1"><nNFSe>123</nNFSe></infNFSe></NFSe>`))
	_ = zw.Close()
	nfseB64 := base64.StdEncoding.EncodeToString(nfseGz.Bytes())

	var gotDPS string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/SefinNacional/nfse" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		raw, _ := base64.StdEncoding.DecodeString(body["dpsXmlGZipB64"])
		zr, _ := gzip.NewReader(bytes.NewReader(raw))
		b, _ := io.ReadAll(zr)
		gotDPS = string(b)
		if strings.Contains(gotDPS, "<nDPS>99</nDPS>") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"erros":[{"Codigo":"E0014","Descricao":"DPS já existe","Complemento":"x"}]}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"chaveAcesso":"35503082227928875000104000000000012326119876543210","idDps":"x","nfseXmlGZipB64":"` + nfseB64 + `","alertas":[]}`))
	}))
	defer srv.Close()

	e := testNationalEmitter(srv.URL + "/SefinNacional/")
	resp, err := postWithClient(e, srv.Client(), cert, "<DPS/>")
	if err != nil {
		t.Fatal(err)
	}
	item, err := e.parseResult(resp, 1, testNationalInvoice())
	if err != nil {
		t.Fatal(err)
	}
	if *item.NFENumber != 123 || len(*item.NFEVerification) != 50 || *item.NFEAmount != 150.5 {
		t.Errorf("unexpected item: %+v", item)
	}
	if gotDPS != "<DPS/>" {
		t.Errorf("server received %q", gotDPS)
	}

	if _, err := postWithClient(e, srv.Client(), cert, "<nDPS>99</nDPS>"); err == nil ||
		!strings.Contains(err.Error(), "[E0014] DPS já existe (x)") {
		t.Errorf("expected E0014 rejection, got %v", err)
	}
}

// postWithClient runs post() against a test server whose CA is only trusted by client.
func postWithClient(e *NationalNfseEmitter, client *http.Client, cert *tls.Certificate, xmlStr string) (*nationalResponse, error) {
	tr := client.Transport.(*http.Transport).Clone()
	tr.TLSClientConfig.Certificates = []tls.Certificate{*cert}
	e.transport = tr
	defer func() { e.transport = nil }()
	return e.post(context.Background(), cert, xmlStr)
}
