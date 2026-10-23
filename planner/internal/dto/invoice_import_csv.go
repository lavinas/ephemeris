package dto

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// CSVInvoiceItem represents an item parsed from a CSV row.
type CSVInvoiceItem struct {
	LineNumber  int
	ServiceID   int64
	ServiceName string
	Description string
	Price       float64
	Quantity    int
}

// CSVGroupedInvoice represents an invoice grouped by (cliente, faturamento, inicio, fim).
type CSVGroupedInvoice struct {
	CustomerID         int64
	CustomerNickname   string
	FirstLineNumber    int
	InvoiceDate        time.Time
	DueDate            time.Time
	PeriodStart        time.Time
	PeriodEnd          time.Time
	ReferenceInvoiceID *int64
	Items              []CSVInvoiceItem
}

// InvoiceImportCSVRequest encapsulates CSV content for importing invoices.
type InvoiceImportCSVRequest struct {
	CSVContent      string               `json:"csv_content,omitempty"`
	Reader          io.Reader            `json:"-"`
	GroupedInvoices []*CSVGroupedInvoice `json:"-"`
}

func (r *InvoiceImportCSVRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		CSVContent string `json:"csv_content"`
		Content    string `json:"content"`
		CSV        string `json:"csv"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.CSVContent != "" {
		r.CSVContent = raw.CSVContent
	} else if raw.Content != "" {
		r.CSVContent = raw.Content
	} else {
		r.CSVContent = raw.CSV
	}
	return nil
}

// Helper function to normalize header names
func normalizeCSVHeader(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	h = strings.ReplaceAll(h, "\ufeff", "") // Strip UTF-8 BOM if present
	h = strings.ReplaceAll(h, "ã", "a")
	h = strings.ReplaceAll(h, "á", "a")
	h = strings.ReplaceAll(h, "à", "a")
	h = strings.ReplaceAll(h, "é", "e")
	h = strings.ReplaceAll(h, "í", "i")
	h = strings.ReplaceAll(h, "ó", "o")
	h = strings.ReplaceAll(h, "ú", "u")
	h = strings.ReplaceAll(h, "ç", "c")
	return h
}

// ParsePrice parses monetary values like "98,00", "110.00", "1.200,50"
func ParsePrice(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	s = strings.TrimPrefix(s, "R$")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("preço vazio")
	}
	if strings.Contains(s, ".") && strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	return strconv.ParseFloat(s, 64)
}

// Validate parses the CSV and performs all validations required by adjust1.md.
// If any error occurs, it returns the description of the first error and does not import anything.
func (r *InvoiceImportCSVRequest) Validate(repo port.Repository) error {
	var rawReader io.Reader
	if r.Reader != nil {
		rawReader = r.Reader
	} else if r.CSVContent != "" {
		rawReader = strings.NewReader(r.CSVContent)
	} else {
		return errors.New("arquivo CSV não informado ou vazio")
	}

	contentBytes, err := io.ReadAll(rawReader)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo CSV: %w", err)
	}

	if len(bytes.TrimSpace(contentBytes)) == 0 {
		return errors.New("arquivo CSV está vazio")
	}

	// Detect delimiter (default to comma, fallback to semicolon)
	delimiter := rune(',')
	firstLine := strings.SplitN(string(contentBytes), "\n", 2)[0]
	if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
		delimiter = ';'
	}

	csvReader := csv.NewReader(bytes.NewReader(contentBytes))
	csvReader.Comma = delimiter
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1

	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("erro ao processar linhas do CSV: %w", err)
	}

	if len(records) < 2 {
		return errors.New("arquivo CSV deve conter pelo menos o cabeçalho e uma linha de dados")
	}

	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		norm := normalizeCSVHeader(col)
		colMap[norm] = i
	}

	// Required columns
	getColIdx := func(keys ...string) (int, bool) {
		for _, k := range keys {
			if idx, ok := colMap[k]; ok {
				return idx, true
			}
		}
		return -1, false
	}

	idxClient, ok1 := getColIdx("cliente", "client", "nickname", "customer")
	idxInvoicing, ok2 := getColIdx("faturamento", "data_faturamento", "data", "invoice_date")
	idxDue, ok3 := getColIdx("vencimento", "data_vencimento", "due_date")
	idxService, ok4 := getColIdx("servico", "service")
	idxStart, ok5 := getColIdx("inicio", "periodo_inicio", "period_start")
	idxEnd, ok6 := getColIdx("fim", "periodo_fim", "period_end")
	idxDesc, _ := getColIdx("descricao", "description")
	idxPrice, ok7 := getColIdx("preco", "price", "valor")
	idxQty, ok8 := getColIdx("quantidade", "quantity", "qtd")
	idxRef, _ := getColIdx("referencia", "reference", "ref", "reference_invoice_id")

	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 || !ok7 || !ok8 {
		return errors.New("cabeçalho CSV inválido: colunas obrigatórias ausentes (cliente, faturamento, vencimento, serviço, inicio, fim, preço, quantidade)")
	}

	// Cache services and customers from repo if available
	var registeredServices []domain.Service
	customerCache := make(map[string]*domain.Customer)

	if repo != nil {
		services, err := repo.FindServices(1)
		if err != nil {
			return fmt.Errorf("erro ao buscar serviços no banco de dados: %w", err)
		}
		registeredServices = services
	}

	groups := make(map[string]*CSVGroupedInvoice)
	var groupOrder []string

	for lineIdx := 1; lineIdx < len(records); lineIdx++ {
		row := records[lineIdx]
		lineNum := lineIdx + 1 // 1-indexed for user-friendly error reporting

		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue // Skip blank lines
		}

		getVal := func(idx int) string {
			if idx >= 0 && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		client := getVal(idxClient)
		if client == "" {
			return fmt.Errorf("linha %d: campo 'cliente' é obrigatório", lineNum)
		}

		faturamentoStr := getVal(idxInvoicing)
		if faturamentoStr == "" {
			return fmt.Errorf("linha %d: campo 'faturamento' é obrigatório", lineNum)
		}
		faturamentoDate, err := ParseFlexibleDate(faturamentoStr)
		if err != nil {
			return fmt.Errorf("linha %d: data de faturamento '%s' inválida: %w", lineNum, faturamentoStr, err)
		}

		vencimentoStr := getVal(idxDue)
		if vencimentoStr == "" {
			return fmt.Errorf("linha %d: campo 'vencimento' é obrigatório", lineNum)
		}
		vencimentoDate, err := ParseFlexibleDate(vencimentoStr)
		if err != nil {
			return fmt.Errorf("linha %d: data de vencimento '%s' inválida: %w", lineNum, vencimentoStr, err)
		}

		if vencimentoDate.Before(faturamentoDate) {
			return fmt.Errorf("linha %d: data de vencimento deve ser maior ou igual à data de faturamento", lineNum)
		}

		inicioStr := getVal(idxStart)
		if inicioStr == "" {
			return fmt.Errorf("linha %d: campo 'inicio' é obrigatório", lineNum)
		}
		inicioDate, err := ParseFlexibleDate(inicioStr)
		if err != nil {
			return fmt.Errorf("linha %d: data de início de período '%s' inválida: %w", lineNum, inicioStr, err)
		}

		fimStr := getVal(idxEnd)
		if fimStr == "" {
			return fmt.Errorf("linha %d: campo 'fim' é obrigatório", lineNum)
		}
		fimDate, err := ParseFlexibleDate(fimStr)
		if err != nil {
			return fmt.Errorf("linha %d: data de fim de período '%s' inválida: %w", lineNum, fimStr, err)
		}

		if fimDate.Before(inicioDate) {
			return fmt.Errorf("linha %d: data de fim de período deve ser maior ou igual à data de início", lineNum)
		}

		servicoStr := getVal(idxService)
		if servicoStr == "" {
			return fmt.Errorf("linha %d: campo 'serviço' é obrigatório", lineNum)
		}

		precoStr := getVal(idxPrice)
		if precoStr == "" {
			return fmt.Errorf("linha %d: campo 'preço' é obrigatório", lineNum)
		}
		preco, err := ParsePrice(precoStr)
		if err != nil || preco <= 0 {
			return fmt.Errorf("linha %d: preço inválido '%s', deve ser maior que zero", lineNum, precoStr)
		}

		qtdStr := getVal(idxQty)
		if qtdStr == "" {
			return fmt.Errorf("linha %d: campo 'quantidade' é obrigatório", lineNum)
		}
		qtd, err := strconv.Atoi(qtdStr)
		if err != nil || qtd <= 0 {
			return fmt.Errorf("linha %d: quantidade inválida '%s', deve ser maior que zero", lineNum, qtdStr)
		}

		descricao := getVal(idxDesc)

		var refID *int64
		refStr := getVal(idxRef)
		if refStr != "" {
			r, err := strconv.ParseInt(refStr, 10, 64)
			if err != nil || r < 0 {
				return fmt.Errorf("linha %d: referência '%s' inválida", lineNum, refStr)
			}
			refID = &r
		}

		// Validate customer existence in DB
		var customerID int64
		if repo != nil {
			cust, ok := customerCache[client]
			if !ok {
				c, err := repo.GetCustomer(1, client)
				if err != nil {
					return fmt.Errorf("linha %d: erro ao consultar cliente '%s': %w", lineNum, client, err)
				}
				if c == nil {
					return fmt.Errorf("linha %d: cliente '%s' não encontrado", lineNum, client)
				}
				customerCache[client] = c
				cust = c
			}
			customerID = cust.ID
		}

		// Validate service existence in DB
		var serviceID int64
		if repo != nil {
			var found bool
			for _, s := range registeredServices {
				if strings.EqualFold(s.Name, servicoStr) {
					serviceID = s.ID
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("linha %d: serviço '%s' não encontrado ou não registrado", lineNum, servicoStr)
			}
		}

		// Group key: client + faturamento + inicio + fim
		groupKey := fmt.Sprintf("%s|%s|%s|%s",
			client,
			faturamentoDate.Format("2006-01-02"),
			inicioDate.Format("2006-01-02"),
			fimDate.Format("2006-01-02"),
		)

		group, exists := groups[groupKey]
		if !exists {
			group = &CSVGroupedInvoice{
				CustomerID:         customerID,
				CustomerNickname:   client,
				FirstLineNumber:    lineNum,
				InvoiceDate:        faturamentoDate,
				DueDate:            vencimentoDate,
				PeriodStart:        inicioDate,
				PeriodEnd:          fimDate,
				ReferenceInvoiceID: refID,
				Items:              make([]CSVInvoiceItem, 0),
			}
			groups[groupKey] = group
			groupOrder = append(groupOrder, groupKey)
		} else {
			// Validate reference consistency across grouped rows
			if refID != nil {
				if group.ReferenceInvoiceID == nil {
					group.ReferenceInvoiceID = refID
				} else if *group.ReferenceInvoiceID != *refID {
					return fmt.Errorf("linha %d: referência divergente na mesma fatura do cliente '%s' (%d vs %d)",
						lineNum, client, *group.ReferenceInvoiceID, *refID)
				}
			}
		}

		group.Items = append(group.Items, CSVInvoiceItem{
			LineNumber:  lineNum,
			ServiceID:   serviceID,
			ServiceName: servicoStr,
			Description: descricao,
			Price:       preco,
			Quantity:    qtd,
		})
	}

	if len(groupOrder) == 0 {
		return errors.New("nenhuma linha válida encontrada para importação no arquivo CSV")
	}

	// Check DB uniqueness for each grouped invoice
	if repo != nil {
		for _, key := range groupOrder {
			g := groups[key]
			exists, err := repo.CheckInvoiceExists(g.CustomerID, g.InvoiceDate, g.PeriodStart, g.PeriodEnd, 0)
			if err != nil {
				return fmt.Errorf("linha %d: erro ao verificar duplicidade de fatura: %w", g.FirstLineNumber, err)
			}
			if exists {
				return fmt.Errorf("linha %d: já existe uma fatura cadastrada para o cliente '%s', faturamento em %s e período %s a %s",
					g.FirstLineNumber,
					g.CustomerNickname,
					g.InvoiceDate.Format("02/01/2006"),
					g.PeriodStart.Format("02/01/2006"),
					g.PeriodEnd.Format("02/01/2006"),
				)
			}
		}
	}

	r.GroupedInvoices = make([]*CSVGroupedInvoice, 0, len(groupOrder))
	for _, key := range groupOrder {
		r.GroupedInvoices = append(r.GroupedInvoices, groups[key])
	}

	return nil
}

// InvoiceImportCSVResponse represents the response after importing a CSV of invoices.
type InvoiceImportCSVResponse struct {
	ResponseBase
	InvoicesImported int     `json:"invoices_imported"`
	ItemsImported    int     `json:"items_imported"`
	TotalAmount      float64 `json:"total_amount"`
}

// NewInvoiceImportCSVResponse creates a new instance of InvoiceImportCSVResponse.
func NewInvoiceImportCSVResponse(statusCode int, statusMessage, errorMessage string, invoicesImported, itemsImported int, totalAmount float64) *InvoiceImportCSVResponse {
	return &InvoiceImportCSVResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		InvoicesImported: invoicesImported,
		ItemsImported:    itemsImported,
		TotalAmount:      totalAmount,
	}
}
