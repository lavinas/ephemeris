package driven

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"billing/internal/domain"
)

type dummyLogger struct{}

func (d *dummyLogger) IPrintf(verbosity int, format string, v ...any) {}
func (d *dummyLogger) Close()                                         {}

func TestSPNfseEmitter_BuildPedidoEnvioRPSXML(t *testing.T) {
	emitter := NewSPNfseEmitter(
		&dummyLogger{},
		"homologacao",
		"",
		"",
		12345,
		0.05,
		"TESTE",
		10,
		"https://nfe.prefeitura.sp.gov.br/ws/lotenfe.asmx",
	)

	doc := "42343223858"
	email := "cliente@teste.com"
	nbs := int64(10101010)

	invoice := &domain.Invoice{
		ID:       1,
		Amount:   150.00,
		Vendor: domain.Vendor{
			Document:    "27928875000104",
			TaxDocument: "57278881",
		},
		Customer: domain.Customer{
			Name:     "Cliente Teste & Cia",
			Document: &doc,
			Email:    &email,
		},
		InvoiceItems: []domain.InvoiceItem{
			{
				ID:          1,
				Description: "Serviço Teste",
				Quantity:    1,
				Price:       150.00,
				NBSCode:     &nbs,
			},
		},
	}

	xmlContent, err := emitter.buildPedidoEnvioRPSXML(invoice, 1001, time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC), "FAKE_SIG")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(xmlContent, `<ValorFinalCobrado>150.00</ValorFinalCobrado>`) {
		t.Errorf("expected <ValorFinalCobrado>150.00</ValorFinalCobrado>, but not found in:\n%s", xmlContent)
	}
	if strings.Contains(xmlContent, `ValorServicos`) || strings.Contains(xmlContent, `ValorInicialCobrado`) {
		t.Errorf("should NOT contain ValorServicos/ValorInicialCobrado, but found in:\n%s", xmlContent)
	}
	// Layout 2: ValorDeducoes must immediately follow TributacaoRPS
	if !strings.Contains(xmlContent, "<TributacaoRPS>T</TributacaoRPS>\n    <ValorDeducoes>") {
		t.Errorf("expected ValorDeducoes right after TributacaoRPS")
	}
	// Order: Discriminacao < ValorFinalCobrado < NBS < cLocPrestacao < IBSCBS
	order := []string{"<Discriminacao>", "<ValorFinalCobrado>", "<ValorIPI>", "<ExigibilidadeSuspensa>", "<PagamentoParceladoAntecipado>", "<NBS>", "<cLocPrestacao>", "<IBSCBS>",
		"<finNFSe>", "<indFinal>", "<cIndOp>100301</cIndOp>", "<indDest>", "<valores>", "<trib>", "<gIBSCBS>", "<cClassTrib>"}
	last := -1
	for _, tag := range order {
		idx := strings.Index(xmlContent, tag)
		if idx <= last {
			t.Errorf("tag %s out of order or missing", tag)
		}
		last = idx
	}
	if !strings.Contains(xmlContent, `<Cabecalho xmlns="" Versao="2">`) {
		t.Errorf("expected Versao=2 in Cabecalho, but not found")
	}
	if !strings.Contains(xmlContent, `<NBS>010101010</NBS>`) {
		t.Errorf("expected padded NBS code <NBS>010101010</NBS>, but not found")
	}
	if !strings.Contains(xmlContent, `<cLocPrestacao>3550308</cLocPrestacao>`) {
		t.Errorf("expected cLocPrestacao 3550308, but not found")
	}
	if !strings.Contains(xmlContent, `<cClassTrib>499001</cClassTrib>`) {
		t.Errorf("expected cClassTrib 499001, but not found")
	}
	if strings.Contains(xmlContent, `<CST>`) {
		t.Errorf("PMSP v02 gIBSCBS has no CST element, but found")
	}
	// PF tomador without address: EnderecoTomador must be omitted (empty Cidade/UF/CEP fail schema)
	if strings.Contains(xmlContent, `EnderecoTomador`) {
		t.Errorf("EnderecoTomador should be omitted, but found in:\n%s", xmlContent)
	}
}

func TestSPNfseEmitter_BuildAndSignRPS(t *testing.T) {
	emitter := NewSPNfseEmitter(
		&dummyLogger{},
		"homologacao",
		"",
		"",
		2692,
		0.05,
		"A",
		10,
		"",
	)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	doc := "42343223858"
	nbs := int64(10101010)

	invoice := &domain.Invoice{
		ID:     1,
		Amount: 250.50,
		Vendor: domain.Vendor{
			Document:    "27928875000104",
			TaxDocument: "57278881", // 8 digits
		},
		Customer: domain.Customer{
			Name:     "Cliente Teste",
			Document: &doc,
		},
		InvoiceItems: []domain.InvoiceItem{
			{NBSCode: &nbs, Price: 250.50, Quantity: 1},
		},
	}

	sig, err := emitter.buildAndSignRPS(invoice, 42, time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC), key)
	if err != nil {
		t.Fatalf("unexpected error signing RPS: %v", err)
	}

	if sig == "" {
		t.Errorf("signature should not be empty")
	}
}
