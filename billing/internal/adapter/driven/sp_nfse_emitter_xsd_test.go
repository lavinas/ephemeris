package driven

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"billing/internal/domain"
)

// pmspSchemaDir holds the official PMSP Layout 2 XSDs
// (downloaded from https://notadomilhao.sf.prefeitura.sp.gov.br/schemas-reformatributaria-v02-7/).
const pmspSchemaDir = "testdata/pmsp_v02"

// xsdValidator returns a function that validates an XML file against an XSD file
// using an external validator (xmllint or python3+xmlschema). Returns nil if none is available.
func xsdValidator() func(xsdPath, xmlPath string) ([]byte, error) {
	if _, err := exec.LookPath("xmllint"); err == nil {
		return func(xsdPath, xmlPath string) ([]byte, error) {
			return exec.Command("xmllint", "--noout", "--schema", xsdPath, xmlPath).CombinedOutput()
		}
	}
	python := os.Getenv("PMSP_XSD_PYTHON") // optional: python interpreter with xmlschema installed
	if python == "" {
		python = "python3"
	}
	if err := exec.Command(python, "-c", "import xmlschema").Run(); err == nil {
		script := `import sys, xmlschema
s = xmlschema.XMLSchema(sys.argv[1])
errs = list(s.iter_errors(sys.argv[2]))
for e in errs:
    print(f"{e.path}: {e.reason}")
sys.exit(1 if errs else 0)`
		return func(xsdPath, xmlPath string) ([]byte, error) {
			return exec.Command(python, "-c", script, xsdPath, xmlPath).CombinedOutput()
		}
	}
	return nil
}

func selfSignedCert(t *testing.T) (*tls.Certificate, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "TESTE NFSE"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	cert, _ := x509.ParseCertificate(der)
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: cert}, key
}

// TestSPNfseEmitter_PedidoEnvioRPS_ValidatesAgainstPMSPSchema builds and signs a full
// PedidoEnvioRPS exactly as sent to the city hall and validates it against the official XSD.
func TestSPNfseEmitter_PedidoEnvioRPS_ValidatesAgainstPMSPSchema(t *testing.T) {
	validate := xsdValidator()
	if validate == nil {
		t.Skip("nenhum validador XSD disponível (instale xmllint/libxml2-utils ou python3 com xmlschema)")
	}

	emitter := NewSPNfseEmitter(&dummyLogger{}, "homologacao", "", "", 5762, 0.0, "A", 10, "")

	doc := "42343223858"
	email := "cliente@teste.com"
	nbs := int64(122051100)
	invoice := &domain.Invoice{
		ID:     1,
		Amount: 150.00,
		Vendor: domain.Vendor{Document: "27928875000104", TaxDocument: "57278881"},
		Customer: domain.Customer{
			Name:     "Cliente Teste & Cia",
			Document: &doc,
			Email:    &email,
		},
		InvoiceItems: []domain.InvoiceItem{
			{ID: 1, Description: "Sessão de atendimento", Quantity: 1, Price: 150.00, NBSCode: &nbs},
		},
	}

	tlsCert, key := selfSignedCert(t)
	emission := time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)

	rpsSig, err := emitter.buildAndSignRPS(invoice, 1001, emission, key)
	if err != nil {
		t.Fatalf("buildAndSignRPS: %v", err)
	}
	xmlContent, err := emitter.buildPedidoEnvioRPSXML(invoice, 1001, emission, rpsSig)
	if err != nil {
		t.Fatalf("buildPedidoEnvioRPSXML: %v", err)
	}
	signed, err := emitter.signXML(xmlContent, tlsCert, tlsCert.Leaf)
	if err != nil {
		t.Fatalf("signXML: %v", err)
	}

	xmlPath := filepath.Join(t.TempDir(), "PedidoEnvioRPS.xml")
	if err := os.WriteFile(xmlPath, []byte(signed), 0o644); err != nil {
		t.Fatalf("write xml: %v", err)
	}
	xsdPath, _ := filepath.Abs(filepath.Join(pmspSchemaDir, "PedidoEnvioRPS_v02.xsd"))

	if out, err := validate(xsdPath, xmlPath); err != nil {
		t.Fatalf("XML não compatível com o schema PMSP v02:\n%s\n\nXML gerado:\n%s", out, signed)
	}
}

