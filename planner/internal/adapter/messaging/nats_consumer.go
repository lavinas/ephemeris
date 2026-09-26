package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"planner/internal/dto"
	"planner/internal/port"
	"planner/internal/service"

	"github.com/nats-io/nats.go"
)

const (
	SubjectCustomerCreated = "customer.created"
	SubjectCustomerUpdated = "customer.updated"
	SubjectVendorCreated   = "vendor.created"
	SubjectVendorUpdated   = "vendor.updated"
	DefaultQueueGroup      = "planner-customer-sync"
)

// CustomerEvent represents an event received from the message broker.
type CustomerEvent struct {
	EventType string            `json:"event_type"`
	Timestamp time.Time         `json:"timestamp"`
	Vendor    string            `json:"vendor"`
	Data      CustomerEventData `json:"data"`
}

// CustomerEventData represents the payload of customer details in the event.
type CustomerEventData struct {
	Name     string  `json:"name"`
	Nickname string  `json:"nickname"`
	Document *string `json:"document,omitempty"`
	Email    *string `json:"email,omitempty"`
	Whatsapp *string `json:"whatsapp,omitempty"`
	Status   *int    `json:"status,omitempty"`
}

// VendorEvent represents an event received for vendor lifecycle.
type VendorEvent struct {
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Vendor    string          `json:"vendor"`
	Data      VendorEventData `json:"data"`
}

// VendorEventData represents the payload of vendor details in the event.
type VendorEventData struct {
	Nickname      string `json:"nickname"`
	LegalName     string `json:"legal_name"`
	TradingName   string `json:"trading_name"`
	Document      string `json:"document"`
	TaxDocument   string `json:"tax_document"`
	AccountBank   string `json:"account_bank"`
	AccountAgency string `json:"account_agency"`
	AccountNumber string `json:"account_number"`
	PixToken      string `json:"pix_token"`
	PixName       string `json:"pix_name"`
	PixCity       string `json:"pix_city"`
	LogoName      string `json:"logo_name"`
	Email         string `json:"email"`
	Whatsapp      string `json:"whatsapp"`
	LastRps       int64  `json:"last_rps"`
	SmtpHost      string `json:"smtp_host"`
	SmtpPort      int    `json:"smtp_port"`
	SmtpUser      string `json:"smtp_user"`
	SmtpPassword  string `json:"smtp_password"`
}

// NATSConsumer is a driver adapter that consumes customer and vendor events from NATS and delegates to planner services.
type NATSConsumer struct {
	url        string
	queueGroup string
	repo       port.Repository
	logger     port.Logger
	nc         *nats.Conn
	subs       []*nats.Subscription
}

// NewNATSConsumer creates a new instance of NATSConsumer.
func NewNATSConsumer(url, queueGroup string, repo port.Repository, logger port.Logger) *NATSConsumer {
	if queueGroup == "" {
		queueGroup = DefaultQueueGroup
	}
	return &NATSConsumer{
		url:        url,
		queueGroup: queueGroup,
		repo:       repo,
		logger:     logger,
	}
}

// Start connects to NATS and subscribes to customer and vendor events.
func (c *NATSConsumer) Start(ctx context.Context) error {
	opts := []nats.Option{
		nats.Name("planner-consumer"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if c.logger != nil && err != nil {
				c.logger.IPrintf(1, "NATS consumer disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			if c.logger != nil {
				c.logger.IPrintf(0, "NATS consumer reconnected to %s", nc.ConnectedUrl())
			}
		}),
	}

	nc, err := nats.Connect(c.url, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS at %s: %w", c.url, err)
	}
	c.nc = nc

	if c.logger != nil {
		c.logger.IPrintf(0, "NATS consumer connected to %s", c.url)
	}

	// Queue subscribe to customer.created
	subCreated, err := nc.QueueSubscribe(SubjectCustomerCreated, c.queueGroup, func(msg *nats.Msg) {
		c.handleCustomerCreated(msg)
	})
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to subscribe to %s: %w", SubjectCustomerCreated, err)
	}
	c.subs = append(c.subs, subCreated)

	// Queue subscribe to customer.updated
	subUpdated, err := nc.QueueSubscribe(SubjectCustomerUpdated, c.queueGroup, func(msg *nats.Msg) {
		c.handleCustomerUpdated(msg)
	})
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to subscribe to %s: %w", SubjectCustomerUpdated, err)
	}
	c.subs = append(c.subs, subUpdated)

	// Queue subscribe to vendor.created
	subVendorCreated, err := nc.QueueSubscribe(SubjectVendorCreated, c.queueGroup, func(msg *nats.Msg) {
		c.handleVendorCreated(msg)
	})
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to subscribe to %s: %w", SubjectVendorCreated, err)
	}
	c.subs = append(c.subs, subVendorCreated)

	// Queue subscribe to vendor.updated
	subVendorUpdated, err := nc.QueueSubscribe(SubjectVendorUpdated, c.queueGroup, func(msg *nats.Msg) {
		c.handleVendorUpdated(msg)
	})
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to subscribe to %s: %w", SubjectVendorUpdated, err)
	}
	c.subs = append(c.subs, subVendorUpdated)

	if c.logger != nil {
		c.logger.IPrintf(0, "NATS consumer listening on [%s, %s, %s, %s] with queue group '%s'",
			SubjectCustomerCreated, SubjectCustomerUpdated, SubjectVendorCreated, SubjectVendorUpdated, c.queueGroup)
	}

	return nil
}

// handleCustomerCreated processes customer.created messages.
func (c *NATSConsumer) handleCustomerCreated(msg *nats.Msg) {
	if c.logger != nil {
		c.logger.IPrintf(0, "Received %s message: %s", SubjectCustomerCreated, string(msg.Data))
	}

	var event CustomerEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		if c.logger != nil {
			c.logger.IPrintf(2, "Failed to decode customer.created event: %v", err)
		}
		return
	}

	var doc, email, whatsapp string
	if event.Data.Document != nil {
		doc = *event.Data.Document
	}
	if event.Data.Email != nil {
		email = *event.Data.Email
	}
	if event.Data.Whatsapp != nil {
		whatsapp = *event.Data.Whatsapp
	}

	createReq := &dto.CustomerCreateRequest{
		Vendor: event.Vendor,
		Items: []*dto.CustomerCreateRequestItem{
			{
				Name:     event.Data.Name,
				Nickname: event.Data.Nickname,
				Document: doc,
				Email:    email,
				Whatsapp: whatsapp,
			},
		},
	}

	createSvc := service.NewCustomerCreate(c.repo, c.logger)
	out := createSvc.Run(createReq)

	if out.GetStatusCode() == 200 {
		if c.logger != nil {
			c.logger.IPrintf(0, "Successfully created customer %s for vendor %s", event.Data.Nickname, event.Vendor)
		}
		return
	}

	// If creation failed (for example, customer already exists in planner), attempt update as fallback (upsert)
	if c.logger != nil {
		c.logger.IPrintf(1, "Customer create returned status %d. Attempting update fallback for %s.",
			out.GetStatusCode(), event.Data.Nickname)
	}

	updateReq := &dto.CustomerUpdateRequest{
		Vendor:   event.Vendor,
		Nickname: event.Data.Nickname,
		Name:     &event.Data.Name,
		Document: event.Data.Document,
		Email:    event.Data.Email,
		Whatsapp: event.Data.Whatsapp,
		Status:   event.Data.Status,
	}

	updateSvc := service.NewCustomerUpdate(c.repo, c.logger)
	updateOut := updateSvc.Run(updateReq)
	if updateOut.GetStatusCode() == 200 && c.logger != nil {
		c.logger.IPrintf(0, "Fallback update succeeded for customer %s", event.Data.Nickname)
	}
}

// handleCustomerUpdated processes customer.updated messages.
func (c *NATSConsumer) handleCustomerUpdated(msg *nats.Msg) {
	if c.logger != nil {
		c.logger.IPrintf(0, "Received %s message: %s", SubjectCustomerUpdated, string(msg.Data))
	}

	var event CustomerEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		if c.logger != nil {
			c.logger.IPrintf(2, "Failed to decode customer.updated event: %v", err)
		}
		return
	}

	updateReq := &dto.CustomerUpdateRequest{
		Vendor:   event.Vendor,
		Nickname: event.Data.Nickname,
		Name:     &event.Data.Name,
		Document: event.Data.Document,
		Email:    event.Data.Email,
		Whatsapp: event.Data.Whatsapp,
		Status:   event.Data.Status,
	}

	updateSvc := service.NewCustomerUpdate(c.repo, c.logger)
	out := updateSvc.Run(updateReq)

	if out.GetStatusCode() == 200 {
		if c.logger != nil {
			c.logger.IPrintf(0, "Successfully updated customer %s for vendor %s", event.Data.Nickname, event.Vendor)
		}
		return
	}

	// If customer does not exist in planner yet, attempt creation as fallback
	if c.logger != nil {
		c.logger.IPrintf(1, "Customer update returned status %d. Attempting create fallback for %s.",
			out.GetStatusCode(), event.Data.Nickname)
	}

	var doc, email, whatsapp string
	if event.Data.Document != nil {
		doc = *event.Data.Document
	}
	if event.Data.Email != nil {
		email = *event.Data.Email
	}
	if event.Data.Whatsapp != nil {
		whatsapp = *event.Data.Whatsapp
	}

	createReq := &dto.CustomerCreateRequest{
		Vendor: event.Vendor,
		Items: []*dto.CustomerCreateRequestItem{
			{
				Name:     event.Data.Name,
				Nickname: event.Data.Nickname,
				Document: doc,
				Email:    email,
				Whatsapp: whatsapp,
			},
		},
	}

	createSvc := service.NewCustomerCreate(c.repo, c.logger)
	createOut := createSvc.Run(createReq)
	if createOut.GetStatusCode() == 200 && c.logger != nil {
		c.logger.IPrintf(0, "Fallback create succeeded for customer %s", event.Data.Nickname)
	}
}

// handleVendorCreated processes vendor.created messages.
func (c *NATSConsumer) handleVendorCreated(msg *nats.Msg) {
	if c.logger != nil {
		c.logger.IPrintf(0, "Received %s message: %s", SubjectVendorCreated, string(msg.Data))
	}

	var event VendorEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		if c.logger != nil {
			c.logger.IPrintf(2, "Failed to decode vendor.created event: %v", err)
		}
		return
	}

	nickname := event.Data.Nickname
	if nickname == "" {
		nickname = event.Vendor
	}

	createReq := &dto.VendorCreateRequest{
		Nickname:      nickname,
		LegalName:     event.Data.LegalName,
		TradingName:   event.Data.TradingName,
		Document:      event.Data.Document,
		TaxDocument:   event.Data.TaxDocument,
		AccountBank:   event.Data.AccountBank,
		AccountAgency: event.Data.AccountAgency,
		AccountNumber: event.Data.AccountNumber,
		PixToken:      event.Data.PixToken,
		PixName:       event.Data.PixName,
		PixCity:       event.Data.PixCity,
		LogoName:      event.Data.LogoName,
		Email:         event.Data.Email,
		Whatsapp:      event.Data.Whatsapp,
		LastRps:       event.Data.LastRps,
		SmtpHost:      event.Data.SmtpHost,
		SmtpPort:      event.Data.SmtpPort,
		SmtpUser:      event.Data.SmtpUser,
		SmtpPassword:  event.Data.SmtpPassword,
	}

	createSvc := service.NewVendorCreate(c.repo, c.logger)
	out := createSvc.Run(createReq)

	if out.GetStatusCode() == 200 {
		if c.logger != nil {
			c.logger.IPrintf(0, "Successfully created vendor %s", nickname)
		}
		return
	}

	// If creation failed (for example, vendor already exists in planner), attempt update as fallback (upsert)
	if c.logger != nil {
		c.logger.IPrintf(1, "Vendor create returned status %d. Attempting update fallback for %s.",
			out.GetStatusCode(), nickname)
	}

	updateReq := buildVendorUpdateRequest(event)
	updateSvc := service.NewVendorUpdate(c.repo, c.logger)
	updateOut := updateSvc.Run(updateReq)
	if updateOut.GetStatusCode() == 200 && c.logger != nil {
		c.logger.IPrintf(0, "Fallback update succeeded for vendor %s", nickname)
	}
}

// handleVendorUpdated processes vendor.updated messages.
func (c *NATSConsumer) handleVendorUpdated(msg *nats.Msg) {
	if c.logger != nil {
		c.logger.IPrintf(0, "Received %s message: %s", SubjectVendorUpdated, string(msg.Data))
	}

	var event VendorEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		if c.logger != nil {
			c.logger.IPrintf(2, "Failed to decode vendor.updated event: %v", err)
		}
		return
	}

	nickname := event.Data.Nickname
	if nickname == "" {
		nickname = event.Vendor
	}

	updateReq := buildVendorUpdateRequest(event)
	updateSvc := service.NewVendorUpdate(c.repo, c.logger)
	out := updateSvc.Run(updateReq)

	if out.GetStatusCode() == 200 {
		if c.logger != nil {
			c.logger.IPrintf(0, "Successfully updated vendor %s", nickname)
		}
		return
	}

	// If vendor does not exist in planner yet, attempt creation as fallback
	if c.logger != nil {
		c.logger.IPrintf(1, "Vendor update returned status %d. Attempting create fallback for %s.",
			out.GetStatusCode(), nickname)
	}

	createReq := &dto.VendorCreateRequest{
		Nickname:      nickname,
		LegalName:     event.Data.LegalName,
		TradingName:   event.Data.TradingName,
		Document:      event.Data.Document,
		TaxDocument:   event.Data.TaxDocument,
		AccountBank:   event.Data.AccountBank,
		AccountAgency: event.Data.AccountAgency,
		AccountNumber: event.Data.AccountNumber,
		PixToken:      event.Data.PixToken,
		PixName:       event.Data.PixName,
		PixCity:       event.Data.PixCity,
		LogoName:      event.Data.LogoName,
		Email:         event.Data.Email,
		Whatsapp:      event.Data.Whatsapp,
		LastRps:       event.Data.LastRps,
		SmtpHost:      event.Data.SmtpHost,
		SmtpPort:      event.Data.SmtpPort,
		SmtpUser:      event.Data.SmtpUser,
		SmtpPassword:  event.Data.SmtpPassword,
	}

	createSvc := service.NewVendorCreate(c.repo, c.logger)
	createOut := createSvc.Run(createReq)
	if createOut.GetStatusCode() == 200 && c.logger != nil {
		c.logger.IPrintf(0, "Fallback create succeeded for vendor %s", nickname)
	}
}

func buildVendorUpdateRequest(event VendorEvent) *dto.VendorUpdateRequest {
	nickname := event.Data.Nickname
	if nickname == "" {
		nickname = event.Vendor
	}
	updateReq := &dto.VendorUpdateRequest{
		Nickname: nickname,
	}
	if event.Data.LegalName != "" {
		updateReq.LegalName = &event.Data.LegalName
	}
	if event.Data.TradingName != "" {
		updateReq.TradingName = &event.Data.TradingName
	}
	if event.Data.Document != "" {
		updateReq.Document = &event.Data.Document
	}
	if event.Data.TaxDocument != "" {
		updateReq.TaxDocument = &event.Data.TaxDocument
	}
	if event.Data.AccountBank != "" {
		updateReq.AccountBank = &event.Data.AccountBank
	}
	if event.Data.AccountAgency != "" {
		updateReq.AccountAgency = &event.Data.AccountAgency
	}
	if event.Data.AccountNumber != "" {
		updateReq.AccountNumber = &event.Data.AccountNumber
	}
	if event.Data.PixToken != "" {
		updateReq.PixToken = &event.Data.PixToken
	}
	if event.Data.PixName != "" {
		updateReq.PixName = &event.Data.PixName
	}
	if event.Data.PixCity != "" {
		updateReq.PixCity = &event.Data.PixCity
	}
	if event.Data.LogoName != "" {
		updateReq.LogoName = &event.Data.LogoName
	}
	if event.Data.Email != "" {
		updateReq.Email = &event.Data.Email
	}
	if event.Data.Whatsapp != "" {
		updateReq.Whatsapp = &event.Data.Whatsapp
	}
	if event.Data.LastRps > 0 {
		updateReq.LastRps = &event.Data.LastRps
	}
	if event.Data.SmtpHost != "" {
		updateReq.SmtpHost = &event.Data.SmtpHost
	}
	if event.Data.SmtpPort > 0 {
		updateReq.SmtpPort = &event.Data.SmtpPort
	}
	if event.Data.SmtpUser != "" {
		updateReq.SmtpUser = &event.Data.SmtpUser
	}
	if event.Data.SmtpPassword != "" {
		updateReq.SmtpPassword = &event.Data.SmtpPassword
	}
	return updateReq
}

// Close unsubscribes and closes the NATS connection.
func (c *NATSConsumer) Close() error {
	for _, sub := range c.subs {
		if sub != nil && sub.IsValid() {
			sub.Unsubscribe()
		}
	}
	c.subs = nil
	if c.nc != nil && !c.nc.IsClosed() {
		c.nc.Drain()
		c.nc.Close()
	}
	return nil
}
