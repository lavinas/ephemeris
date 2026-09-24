package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/service"

	"github.com/nats-io/nats.go"
)

const (
	SubjectCustomerCreated = "customer.created"
	SubjectCustomerUpdated = "customer.updated"
	DefaultQueueGroup      = "billing-customer-sync"
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

// NATSConsumer is a driver adapter that consumes customer events from NATS and delegates to billing services.
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

// Start connects to NATS and subscribes to customer events.
func (c *NATSConsumer) Start(ctx context.Context) error {
	opts := []nats.Option{
		nats.Name("billing-consumer"),
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

	if c.logger != nil {
		c.logger.IPrintf(0, "NATS consumer listening on [%s, %s] with queue group '%s'",
			SubjectCustomerCreated, SubjectCustomerUpdated, c.queueGroup)
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

	// If creation failed (for example, customer already exists in billing), attempt update as fallback (upsert)
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

	// If customer does not exist in billing yet, attempt creation as fallback
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
