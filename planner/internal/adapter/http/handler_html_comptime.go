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
	ID                            int64
	VendorID                      int64
	CustomerID                    int64
	CustomerNickname              string
	CustomerName                  string
	ServiceID                     int64
	ServiceName                   string
	ServiceDescription            string
	ServiceMinutes                int
	Month                         time.Time
	MonthFormatado                string
	InvoicedMinutes               int
	ConsumedMinutes               int
	AdjustedMinutes               int
	AdjustDescription             string
	PredictMinutes                int
	DoneMinutes                   int
	MonthBalance                  int
	Balance                       int
	PredictSessions               float64
	DoneSessions                  float64
	AdjustedSessions              float64
	MonthBalanceSessions          float64
	BalanceSessions               float64
	PredictSessionsFormatado      string
	DoneSessionsFormatado         string
	AdjustedSessionsFormatado     string
	MonthBalanceSessionsFormatado string
	BalanceSessionsFormatado      string
	UpdatedAt                     time.Time
	UpdatedAtFormatado            string
}

func buildComptimeRecordItem(rec port.ComptimeView) comptimeRecordItem {
	serviceMinutes := rec.ServiceMinutes
	predictMinutes := rec.PredictMinutes
	doneMinutes := rec.DoneMinutes
	adjustedMinutes := rec.AdjustedMinutes
	monthBalance := rec.MonthBalance
	balance := rec.Balance
	predictSessions := rec.PredictSessions
	doneSessions := rec.DoneSessions
	adjustedSessions := rec.AdjustedSessions
	monthBalanceSessions := rec.MonthBalanceSessions
	balanceSessions := rec.BalanceSessions

	// Fallback caso a view venha diretamente de um repositório sem passar pelo ComptimeService
	if serviceMinutes <= 0 {
		serviceMinutes = 60
		predictMinutes = rec.InvoicedMinutes
		doneMinutes = rec.ConsumedMinutes
		monthBalance = predictMinutes - doneMinutes + adjustedMinutes
		predictSessions = domain.CalculateSessionsFromMinutes(predictMinutes, serviceMinutes)
		doneSessions = domain.CalculateSessionsFromMinutes(doneMinutes, serviceMinutes)
		adjustedSessions = domain.CalculateSessionsFromMinutes(adjustedMinutes, serviceMinutes)
		monthBalanceSessions = domain.CalculateSessionsFromMinutes(monthBalance, serviceMinutes)
		balanceSessions = domain.CalculateSessionsFromMinutes(balance, serviceMinutes)
	} else if adjustedSessions == 0 && adjustedMinutes != 0 {
		adjustedSessions = domain.CalculateSessionsFromMinutes(adjustedMinutes, serviceMinutes)
	}

	return comptimeRecordItem{
		ID:                            rec.ID,
		VendorID:                      rec.VendorID,
		CustomerID:                    rec.CustomerID,
		CustomerNickname:              rec.CustomerNickname,
		CustomerName:                  rec.CustomerName,
		ServiceID:                     rec.ServiceID,
		ServiceName:                   rec.ServiceName,
		ServiceDescription:            rec.ServiceDescription,
		ServiceMinutes:                serviceMinutes,
		Month:                         rec.Month,
		MonthFormatado:                rec.Month.Format("2006-01"),
		InvoicedMinutes:               rec.InvoicedMinutes,
		ConsumedMinutes:               rec.ConsumedMinutes,
		AdjustedMinutes:               adjustedMinutes,
		AdjustDescription:             rec.AdjustDescription,
		PredictMinutes:                predictMinutes,
		DoneMinutes:                   doneMinutes,
		MonthBalance:                  monthBalance,
		Balance:                       balance,
		PredictSessions:               predictSessions,
		DoneSessions:                  doneSessions,
		AdjustedSessions:              adjustedSessions,
		MonthBalanceSessions:          monthBalanceSessions,
		BalanceSessions:               balanceSessions,
		PredictSessionsFormatado:      fmt.Sprintf("%.2f", predictSessions),
		DoneSessionsFormatado:         fmt.Sprintf("%.2f", doneSessions),
		AdjustedSessionsFormatado:     fmt.Sprintf("%.2f", adjustedSessions),
		MonthBalanceSessionsFormatado: fmt.Sprintf("%.2f", monthBalanceSessions),
		BalanceSessionsFormatado:      fmt.Sprintf("%.2f", balanceSessions),
		UpdatedAt:                     rec.UpdatedAt,
		UpdatedAtFormatado:            rec.UpdatedAt.Format("02/01/2006 15:04"),
	}
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

	for i, rec := range records {
		items[i] = buildComptimeRecordItem(rec)
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

// ComptimeAdjustModal renders the modal for creating, updating or zeroing/deleting a manual hour adjustment.
func (h *HandlerHtml) ComptimeAdjustModal(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if _, ok := h.repo.(port.ComptimeRepository); !ok {
		http.Error(w, "Repository não disponível", http.StatusInternalServerError)
		return
	}

	item, ok := h.findSingleComptimeRecordItem(r, id)
	if !ok {
		http.Error(w, "Registro não encontrado", http.StatusNotFound)
		return
	}

	h.tmplComptime.ExecuteTemplate(w, "modal_ajustar", item)
}

// ComptimeAdjust handles creating, updating or zeroing a manual hour adjustment and re-rendering the updated table.
func (h *HandlerHtml) ComptimeAdjust(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		idStr = r.FormValue("comptime_id")
	}
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	adjMinStr := strings.TrimSpace(r.FormValue("adjusted_minutes"))
	adjustedMinutes := 0
	if adjMinStr != "" {
		val, errParse := strconv.Atoi(adjMinStr)
		if errParse != nil {
			http.Error(w, "Minutos ajustados inválidos", http.StatusBadRequest)
			return
		}
		adjustedMinutes = val
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if adjustedMinutes != 0 && description == "" {
		http.Error(w, "Justificativa/descrição é obrigatória para ajustes diferentes de zero", http.StatusBadRequest)
		return
	}

	comptimeSvc := h.getComptimeService()
	if comptimeSvc == nil {
		http.Error(w, "ComptimeService not available", http.StatusInternalServerError)
		return
	}

	if _, _, err := comptimeSvc.SaveComptimeAdjustByID(r.Context(), id, adjustedMinutes, description); err != nil {
		h.logger.IPrintf(1, "Erro ao ajustar comptime %d: %v", id, err)
		http.Error(w, fmt.Sprintf("Erro ao ajustar: %v", err), http.StatusBadRequest)
		return
	}

	data := h.buildComptimePageData(r)
	h.tmplComptime.ExecuteTemplate(w, "tabela", data)
}

// ComptimeAdjustDelete handles deleting/zeroing a manual hour adjustment and re-rendering the updated table.
func (h *HandlerHtml) ComptimeAdjustDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		idStr = r.FormValue("comptime_id")
	}
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

	if _, err := comptimeSvc.DeleteComptimeAdjustByID(r.Context(), id); err != nil {
		h.logger.IPrintf(1, "Erro ao remover ajuste do comptime %d: %v", id, err)
		http.Error(w, fmt.Sprintf("Erro ao remover ajuste: %v", err), http.StatusBadRequest)
		return
	}

	data := h.buildComptimePageData(r)
	h.tmplComptime.ExecuteTemplate(w, "tabela", data)
}

func (h *HandlerHtml) findSingleComptimeRecordItem(r *http.Request, id int64) (comptimeRecordItem, bool) {
	cr, ok := h.repo.(port.ComptimeRepository)
	if !ok {
		return comptimeRecordItem{}, false
	}

	comptime, err := cr.GetComptime(r.Context(), id)
	if err != nil || comptime == nil {
		return comptimeRecordItem{}, false
	}

	if comptimeSvc := h.getComptimeService(); comptimeSvc != nil {
		m := domain.NormalizeMonth(comptime.Month)
		views, _, errList := comptimeSvc.ListComptime(r.Context(), port.ComptimeFilter{
			VendorID:   comptime.VendorID,
			CustomerID: &comptime.CustomerID,
			ServiceID:  &comptime.ServiceID,
			StartMonth: &m,
			EndMonth:   &m,
		})
		if errList == nil {
			for _, v := range views {
				if v.ID == id {
					if v.AdjustDescription == "" {
						if ca, errAdj := cr.FindComptimeAdjustByComptimeID(r.Context(), id); errAdj == nil && ca != nil {
							v.AdjustDescription = ca.Description
						}
					}
					return buildComptimeRecordItem(v), true
				}
			}
		}
	}

	customerName := fmt.Sprintf("Cliente #%d", comptime.CustomerID)
	customerNickname := ""
	if cust, err := h.repo.GetCustomerByID(comptime.CustomerID); err == nil && cust != nil {
		customerName = cust.Name
		customerNickname = cust.Nickname
	}

	serviceName := fmt.Sprintf("Serviço #%d", comptime.ServiceID)
	serviceDescription := ""
	serviceMinutes := 60
	if svc, err := h.repo.GetServiceByID(comptime.ServiceID); err == nil && svc != nil {
		serviceName = svc.Name
		if svc.Description != nil {
			serviceDescription = *svc.Description
		}
		if svc.SessionMinutes != nil && *svc.SessionMinutes > 0 {
			serviceMinutes = *svc.SessionMinutes
		}
	}

	adjustDescription := ""
	if ca, errAdj := cr.FindComptimeAdjustByComptimeID(r.Context(), comptime.ID); errAdj == nil && ca != nil {
		adjustDescription = ca.Description
	}

	return buildComptimeRecordItem(port.ComptimeView{
		ID:                   comptime.ID,
		VendorID:             comptime.VendorID,
		CustomerID:           comptime.CustomerID,
		CustomerNickname:     customerNickname,
		CustomerName:         customerName,
		ServiceID:            comptime.ServiceID,
		ServiceName:          serviceName,
		ServiceDescription:   serviceDescription,
		ServiceMinutes:       serviceMinutes,
		Month:                comptime.Month,
		InvoicedMinutes:      comptime.InvoicedMinutes,
		ConsumedMinutes:      comptime.ConsumedMinutes,
		AdjustedMinutes:      comptime.AdjustedMinutes,
		AdjustDescription:    adjustDescription,
		PredictMinutes:       comptime.InvoicedMinutes,
		DoneMinutes:          comptime.ConsumedMinutes,
		MonthBalance:         comptime.MonthBalance(),
		Balance:              comptime.Balance,
		PredictSessions:      domain.CalculateSessionsFromMinutes(comptime.InvoicedMinutes, serviceMinutes),
		DoneSessions:         domain.CalculateSessionsFromMinutes(comptime.ConsumedMinutes, serviceMinutes),
		AdjustedSessions:     domain.CalculateSessionsFromMinutes(comptime.AdjustedMinutes, serviceMinutes),
		MonthBalanceSessions: domain.CalculateSessionsFromMinutes(comptime.MonthBalance(), serviceMinutes),
		BalanceSessions:      domain.CalculateSessionsFromMinutes(comptime.Balance, serviceMinutes),
		UpdatedAt:            comptime.UpdatedAt,
	}), true
}
