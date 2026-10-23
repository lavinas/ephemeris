package http

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

type comptimeRecordItem struct {
	ID                 int64
	VendorID           int64
	CustomerID         int64
	CustomerNickname   string
	CustomerName       string
	ServiceID          int64
	ServiceName        string
	ServiceDescription string
	Month              time.Time
	MonthFormatado     string
	InvoicedMinutes    int
	ConsumedMinutes    int
	MonthBalance       int
	Balance            int
	UpdatedAt          time.Time
	UpdatedAtFormatado string
}

// ComptimeHTML renders the Banco de Horas HTML view (full index or HTMX fragment).
func (h *HandlerHtml) ComptimeHTML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := h.buildComptimePageData(r)
	if r.Header.Get("HX-Request") == "true" {
		h.tmplComptime.ExecuteTemplate(w, "fragment", data)
		return
	}
	h.tmplComptime.ExecuteTemplate(w, "index", data)
}

// ComptimeTable handles POST /html/comptime/tabela to render the dynamic table rows and pagination.
func (h *HandlerHtml) ComptimeTable(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	data := h.buildComptimePageData(r)
	h.tmplComptime.ExecuteTemplate(w, "tabela", data)
}

// ComptimeTableReset resets all query filters and renders default table.
func (h *HandlerHtml) ComptimeTableReset(w http.ResponseWriter, r *http.Request) {
	r.Form = nil
	r.URL.RawQuery = ""
	data := h.buildComptimePageData(r)
	h.tmplComptime.ExecuteTemplate(w, "tabela", data)
}

// ComptimeResyncModal renders the modal for triggering a resynchronization job.
func (h *HandlerHtml) ComptimeResyncModal(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	firstDayOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	today := now.Format("2006-01-02")

	vendorID := int64(1)
	if vidStr := r.FormValue("vendor_id"); vidStr != "" {
		if vid, err := strconv.ParseInt(vidStr, 10, 64); err == nil && vid > 0 {
			vendorID = vid
		}
	}

	var selectedCustomerID int64
	var selectedNickname string
	if cStr := r.FormValue("customer_id"); cStr != "" {
		if c, err := strconv.ParseInt(cStr, 10, 64); err == nil && c > 0 {
			selectedCustomerID = c
		}
	}
	if nick := strings.TrimSpace(r.FormValue("nickname")); nick != "" {
		selectedNickname = nick
	}
	if selectedCustomerID == 0 && selectedNickname != "" && h.repo != nil {
		if cust, err := h.repo.GetCustomer(vendorID, selectedNickname); err == nil && cust != nil {
			selectedCustomerID = cust.ID
		}
	} else if selectedCustomerID > 0 && selectedNickname == "" && h.repo != nil {
		if cust, err := h.repo.GetCustomerByID(selectedCustomerID); err == nil && cust != nil {
			selectedNickname = cust.Nickname
		}
	}

	var selectedServiceID int64
	if sStr := r.FormValue("service_id"); sStr != "" {
		if s, err := strconv.ParseInt(sStr, 10, 64); err == nil && s > 0 {
			selectedServiceID = s
		}
	}

	var customers []domain.Customer
	if h.repo != nil {
		custs, err := h.repo.FindCustomers(0, 0, vendorID, nil, nil, nil, nil, nil, nil)
		if err == nil {
			customers = custs
		}
	}

	var services []domain.Service
	if h.repo != nil {
		svcs, err := h.repo.FindServices(vendorID)
		if err == nil {
			services = svcs
		}
	}

	data := map[string]interface{}{
		"DefaultStartDate":   firstDayOfYear,
		"DefaultEndDate":     today,
		"SelectedCustomerID": selectedCustomerID,
		"SelectedNickname":   selectedNickname,
		"SelectedServiceID":  selectedServiceID,
		"Nicknames":          h.getCustomersNicknames(),
		"Customers":          customers,
		"Services":           services,
	}

	h.tmplComptime.ExecuteTemplate(w, "modal_resync", data)
}

// ComptimeResync executes the asynchronous or synchronous resync job and displays metrics.
func (h *HandlerHtml) ComptimeResync(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	startStr := r.FormValue("start_date")
	endStr := r.FormValue("end_date")
	dataType := r.FormValue("data_type")
	if dataType == "" {
		dataType = "todos"
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		startDate = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		endDate = time.Now()
	}

	vendorID := int64(1)
	if vidStr := r.FormValue("vendor_id"); vidStr != "" {
		if vid, err := strconv.ParseInt(vidStr, 10, 64); err == nil && vid > 0 {
			vendorID = vid
		}
	}

	var customerID *int64
	customerName := "Todos os Clientes"
	if cStr := r.FormValue("customer_id"); cStr != "" {
		if c, err := strconv.ParseInt(cStr, 10, 64); err == nil && c > 0 {
			customerID = &c
			if h.repo != nil {
				if cust, err := h.repo.GetCustomerByID(c); err == nil && cust != nil {
					customerName = cust.Name
					if cust.Nickname != "" {
						customerName = cust.Nickname + " - " + cust.Name
					}
				}
			}
		}
	}
	if customerID == nil {
		if nick := strings.TrimSpace(r.FormValue("nickname")); nick != "" && h.repo != nil {
			if cust, err := h.repo.GetCustomer(vendorID, nick); err == nil && cust != nil {
				customerID = &cust.ID
				customerName = cust.Name
				if cust.Nickname != "" {
					customerName = cust.Nickname + " - " + cust.Name
				}
			}
		}
	}

	var serviceID *int64
	serviceName := "Todos os Serviços"
	if sStr := r.FormValue("service_id"); sStr != "" {
		if s, err := strconv.ParseInt(sStr, 10, 64); err == nil && s > 0 {
			serviceID = &s
			if h.repo != nil {
				if svc, err := h.repo.GetServiceByID(s); err == nil && svc != nil {
					serviceName = svc.Name
				}
			}
		}
	}

	comptimeSvc := h.getComptimeService()
	if comptimeSvc == nil {
		http.Error(w, "ComptimeService not available", http.StatusInternalServerError)
		return
	}

	result, err := comptimeSvc.Resync(r.Context(), vendorID, startDate, endDate, dataType, customerID, serviceID)
	if err != nil {
		h.logger.IPrintf(1, "Erro no job de ressincronização do banco de horas: %v", err)
		http.Error(w, fmt.Sprintf("Erro ao ressincronizar: %v", err), http.StatusInternalServerError)
		return
	}

	durationFormatted := fmt.Sprintf("%v", result.Duration.Round(time.Millisecond))
	modalData := map[string]interface{}{
		"CustomerName":    customerName,
		"ServiceName":     serviceName,
		"SessionsSynced":  result.SessionsSynced,
		"InvoicesSynced":  result.InvoicesSynced,
		"ComptimeUpdated": result.ComptimeUpdated,
		"Duration":        durationFormatted,
	}

	h.tmplComptime.ExecuteTemplate(w, "modal_resync_resultado", modalData)
}

// buildComptimePageData queries comptime records and builds context map for template.
func (h *HandlerHtml) buildComptimePageData(r *http.Request) map[string]interface{} {
	page := 1
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	} else if pStr := r.FormValue("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 15

	nickname := r.FormValue("nickname")
	if nickname == "" {
		nickname = r.URL.Query().Get("nickname")
	}

	var customerID *int64
	if cStr := r.FormValue("customer_id"); cStr != "" {
		if c, err := strconv.ParseInt(cStr, 10, 64); err == nil && c > 0 {
			customerID = &c
		}
	}

	var serviceID *int64
	if sStr := r.FormValue("service_id"); sStr != "" {
		if s, err := strconv.ParseInt(sStr, 10, 64); err == nil && s > 0 {
			serviceID = &s
		}
	}

	var startMonth *time.Time
	mesInicioStr := r.FormValue("mes_inicio")
	if mesInicioStr != "" {
		if t, err := time.Parse("2006-01", mesInicioStr); err == nil {
			norm := domain.NormalizeMonth(t)
			startMonth = &norm
		} else if t, err := time.Parse("2006-01-02", mesInicioStr); err == nil {
			norm := domain.NormalizeMonth(t)
			startMonth = &norm
		}
	}

	var endMonth *time.Time
	mesFimStr := r.FormValue("mes_fim")
	if mesFimStr != "" {
		if t, err := time.Parse("2006-01", mesFimStr); err == nil {
			norm := domain.NormalizeMonth(t)
			endMonth = &norm
		} else if t, err := time.Parse("2006-01-02", mesFimStr); err == nil {
			norm := domain.NormalizeMonth(t)
			endMonth = &norm
		}
	}

	filter := port.ComptimeFilter{
		VendorID:         1,
		CustomerID:       customerID,
		CustomerNickname: nickname,
		ServiceID:        serviceID,
		StartMonth:       startMonth,
		EndMonth:         endMonth,
		Page:             page,
		PageSize:         pageSize,
	}

	var records []port.ComptimeView
	var totalRegistros int64
	comptimeSvc := h.getComptimeService()
	if comptimeSvc != nil {
		views, total, err := comptimeSvc.ListComptime(r.Context(), filter)
		if err == nil {
			records = views
			totalRegistros = total
		} else {
			h.logger.IPrintf(1, "Erro ao listar banco de horas via serviço: %v", err)
		}
	} else if cr, ok := h.repo.(port.ComptimeRepository); ok {
		views, total, err := cr.ListComptime(r.Context(), filter)
		if err == nil {
			records = views
			totalRegistros = total
		} else {
			h.logger.IPrintf(1, "Erro ao listar banco de horas: %v", err)
		}
	}

	items := make([]comptimeRecordItem, len(records))
	totalFaturados := 0
	totalConsumidos := 0
	totalSaldoMes := 0
	totalSaldo := 0

	for i, rec := range records {
		monthBalance := rec.MonthBalance
		if monthBalance == 0 && (rec.InvoicedMinutes != 0 || rec.ConsumedMinutes != 0) {
			monthBalance = rec.InvoicedMinutes - rec.ConsumedMinutes
		}
		items[i] = comptimeRecordItem{
			ID:                 rec.ID,
			VendorID:           rec.VendorID,
			CustomerID:         rec.CustomerID,
			CustomerNickname:   rec.CustomerNickname,
			CustomerName:       rec.CustomerName,
			ServiceID:          rec.ServiceID,
			ServiceName:        rec.ServiceName,
			ServiceDescription: rec.ServiceDescription,
			Month:              rec.Month,
			MonthFormatado:     rec.Month.Format("2006-01"),
			InvoicedMinutes:    rec.InvoicedMinutes,
			ConsumedMinutes:    rec.ConsumedMinutes,
			MonthBalance:       monthBalance,
			Balance:            rec.Balance,
			UpdatedAt:          rec.UpdatedAt,
			UpdatedAtFormatado: rec.UpdatedAt.Format("02/01/2006 15:04"),
		}
		totalFaturados += rec.InvoicedMinutes
		totalConsumidos += rec.ConsumedMinutes
		totalSaldoMes += monthBalance
		totalSaldo += rec.Balance
	}

	totalPages := int(math.Ceil(float64(totalRegistros) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	// Fetch customers and services for filters
	var customers []domain.Customer
	if h.repo != nil {
		custs, err := h.repo.FindCustomers(0, 0, 1, nil, nil, nil, nil, nil, nil)
		if err == nil {
			customers = custs
		}
	}

	var services []domain.Service
	if h.repo != nil {
		svcs, err := h.repo.FindServices(1)
		if err == nil {
			services = svcs
		}
	}

	var custFilterID int64
	if customerID != nil {
		custFilterID = *customerID
	}
	var servFilterID int64
	if serviceID != nil {
		servFilterID = *serviceID
	}

	return map[string]interface{}{
		"Registros":        items,
		"Pagina":           page,
		"TotalPaginas":     totalPages,
		"TotalRegistros":   totalRegistros,
		"TotalFaturados":   totalFaturados,
		"TotalConsumidos":  totalConsumidos,
		"TotalSaldoMes":    totalSaldoMes,
		"TotalSaldo":       totalSaldo,
		"FiltroMesInicio":  mesInicioStr,
		"FiltroMesFim":     mesFimStr,
		"FiltroNickname":   nickname,
		"FiltroCustomerID": custFilterID,
		"FiltroServiceID":  servFilterID,
		"Nicknames":        h.getCustomersNicknames(),
		"Customers":        customers,
		"Services":         services,
	}
}

// ComptimeDeleteWarning renders the inline deletion confirmation row for a comptime record.
func (h *HandlerHtml) ComptimeDeleteWarning(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	cr, ok := h.repo.(port.ComptimeRepository)
	if !ok {
		http.Error(w, "Repository não disponível", http.StatusInternalServerError)
		return
	}

	comptime, err := cr.GetComptime(r.Context(), id)
	if err != nil || comptime == nil {
		http.Error(w, "Registro não encontrado", http.StatusNotFound)
		return
	}

	customerName := fmt.Sprintf("Cliente #%d", comptime.CustomerID)
	customerNickname := ""
	if cust, err := h.repo.GetCustomerByID(comptime.CustomerID); err == nil && cust != nil {
		customerName = cust.Name
		customerNickname = cust.Nickname
	}

	serviceName := fmt.Sprintf("Serviço #%d", comptime.ServiceID)
	serviceDescription := ""
	if svc, err := h.repo.GetServiceByID(comptime.ServiceID); err == nil && svc != nil {
		serviceName = svc.Name
		if svc.Description != nil {
			serviceDescription = *svc.Description
		}
	}

	item := comptimeRecordItem{
		ID:                 comptime.ID,
		VendorID:           comptime.VendorID,
		CustomerID:         comptime.CustomerID,
		CustomerNickname:   customerNickname,
		CustomerName:       customerName,
		ServiceID:          comptime.ServiceID,
		ServiceName:        serviceName,
		ServiceDescription: serviceDescription,
		Month:              comptime.Month,
		MonthFormatado:     comptime.Month.Format("2006-01"),
		InvoicedMinutes:    comptime.InvoicedMinutes,
		ConsumedMinutes:    comptime.ConsumedMinutes,
		MonthBalance:       comptime.MonthBalance(),
		Balance:            comptime.Balance,
		UpdatedAt:          comptime.UpdatedAt,
		UpdatedAtFormatado: comptime.UpdatedAt.Format("02/01/2006 15:04"),
	}

	h.tmplComptime.ExecuteTemplate(w, "linha_comptime_deletar_aviso", item)
}

// ComptimeCancelDelete restores the normal row display for a comptime record without deleting.
func (h *HandlerHtml) ComptimeCancelDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	cr, ok := h.repo.(port.ComptimeRepository)
	if !ok {
		http.Error(w, "Repository não disponível", http.StatusInternalServerError)
		return
	}

	comptime, err := cr.GetComptime(r.Context(), id)
	if err != nil || comptime == nil {
		http.Error(w, "Registro não encontrado", http.StatusNotFound)
		return
	}

	customerName := fmt.Sprintf("Cliente #%d", comptime.CustomerID)
	customerNickname := ""
	if cust, err := h.repo.GetCustomerByID(comptime.CustomerID); err == nil && cust != nil {
		customerName = cust.Name
		customerNickname = cust.Nickname
	}

	serviceName := fmt.Sprintf("Serviço #%d", comptime.ServiceID)
	serviceDescription := ""
	if svc, err := h.repo.GetServiceByID(comptime.ServiceID); err == nil && svc != nil {
		serviceName = svc.Name
		if svc.Description != nil {
			serviceDescription = *svc.Description
		}
	}

	item := comptimeRecordItem{
		ID:                 comptime.ID,
		VendorID:           comptime.VendorID,
		CustomerID:         comptime.CustomerID,
		CustomerNickname:   customerNickname,
		CustomerName:       customerName,
		ServiceID:          comptime.ServiceID,
		ServiceName:        serviceName,
		ServiceDescription: serviceDescription,
		Month:              comptime.Month,
		MonthFormatado:     comptime.Month.Format("2006-01"),
		InvoicedMinutes:    comptime.InvoicedMinutes,
		ConsumedMinutes:    comptime.ConsumedMinutes,
		MonthBalance:       comptime.MonthBalance(),
		Balance:            comptime.Balance,
		UpdatedAt:          comptime.UpdatedAt,
		UpdatedAtFormatado: comptime.UpdatedAt.Format("02/01/2006 15:04"),
	}

	h.tmplComptime.ExecuteTemplate(w, "linha_comptime", item)
}

// ComptimeDelete handles deleting a comptime record and re-rendering the updated table.
func (h *HandlerHtml) ComptimeDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	comptimeSvc := h.getComptimeService()
	if comptimeSvc == nil {
		http.Error(w, "ComptimeService not available", http.StatusInternalServerError)
		return
	}

	if err := comptimeSvc.DeleteComptime(r.Context(), id); err != nil {
		h.logger.IPrintf(1, "Erro ao excluir comptime %d: %v", id, err)
		http.Error(w, fmt.Sprintf("Erro ao excluir: %v", err), http.StatusInternalServerError)
		return
	}

	data := h.buildComptimePageData(r)
	h.tmplComptime.ExecuteTemplate(w, "tabela", data)
}
