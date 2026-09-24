package http

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"signup/internal/domain"
	"signup/internal/port"
)

type mockLogger struct {
	mu       sync.Mutex
	Messages []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{Messages: make([]string, 0)}
}

func (m *mockLogger) IPrintf(level int, format string, v ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, fmt.Sprintf(format, v...))
}

func (m *mockLogger) Close() {}

type mockRepository struct {
	SaveFunc                func(model interface{}) error
	FindFunc                func(model interface{}, conditions map[string]interface{}, page, pagesize int, orderBy ...string) ([]interface{}, error)
	BeginTransactionFunc    func() error
	CommitTransactionFunc   func() error
	RollbackTransactionFunc func() error
	CloseFunc               func() error
}

func (m *mockRepository) BeginTransaction() error {
	if m.BeginTransactionFunc != nil {
		return m.BeginTransactionFunc()
	}
	return nil
}

func (m *mockRepository) CommitTransaction() error {
	if m.CommitTransactionFunc != nil {
		return m.CommitTransactionFunc()
	}
	return nil
}

func (m *mockRepository) RollbackTransaction() error {
	if m.RollbackTransactionFunc != nil {
		return m.RollbackTransactionFunc()
	}
	return nil
}

func (m *mockRepository) Save(model interface{}) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(model)
	}
	return nil
}

func (m *mockRepository) Find(model interface{}, conditions map[string]interface{}, page, pagesize int, orderBy ...string) ([]interface{}, error) {
	if m.FindFunc != nil {
		return m.FindFunc(model, conditions, page, pagesize, orderBy...)
	}
	return nil, nil
}

func (m *mockRepository) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// memoryRepo implements an in-memory repository for realistic handler testing
type memoryRepo struct {
	mu        sync.RWMutex
	nextID    int64
	Vendors   map[int64]*domain.Vendor
	Customers map[int64]*domain.Customer
	Users     map[int64]*domain.User
}

func newMemoryRepo() *memoryRepo {
	r := &memoryRepo{
		nextID:    1,
		Vendors:   make(map[int64]*domain.Vendor),
		Customers: make(map[int64]*domain.Customer),
		Users:     make(map[int64]*domain.User),
	}

	// Seed a default test vendor
	defaultVendor := &domain.Vendor{
		DomainBase:  domain.DomainBase{Repo: r},
		ID:          1,
		Nickname:    "acme",
		LegalName:   "Acme Corporation",
		TradingName: "Acme",
		Document:    "11144477735",
		Email:       "contact@acme.com",
		Whatsapp:    "+5511999999999",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.Vendors[defaultVendor.ID] = defaultVendor
	r.nextID = 2
	return r
}

func (m *memoryRepo) BeginTransaction() error    { return nil }
func (m *memoryRepo) CommitTransaction() error   { return nil }
func (m *memoryRepo) RollbackTransaction() error { return nil }
func (m *memoryRepo) Close() error               { return nil }

func (m *memoryRepo) Save(model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch entity := model.(type) {
	case *domain.Vendor:
		if entity.ID == 0 {
			entity.ID = m.nextID
			m.nextID++
		}
		entity.DomainBase = domain.DomainBase{Repo: m}
		m.Vendors[entity.ID] = entity
		return nil

	case *domain.Customer:
		if entity.ID == 0 {
			entity.ID = m.nextID
			m.nextID++
		}
		entity.DomainBase = domain.DomainBase{Repo: m}
		m.Customers[entity.ID] = entity
		return nil

	case *domain.User:
		if entity.ID == 0 {
			entity.ID = m.nextID
			m.nextID++
		}
		entity.DomainBase = domain.DomainBase{Repo: m}
		m.Users[entity.ID] = entity
		return nil

	default:
		return fmt.Errorf("unsupported model type in mock Save: %T", model)
	}
}

func (m *memoryRepo) Find(model interface{}, conditions map[string]interface{}, page, pagesize int, orderBy ...string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []interface{}

	switch model.(type) {
	case *domain.Vendor:
		for _, v := range m.Vendors {
			match := true
			if nick, ok := conditions["nickname = ?"]; ok {
				if v.Nickname != nick {
					match = false
				}
			}
			if id, ok := conditions["id = ?"]; ok {
				if v.ID != id {
					match = false
				}
			}
			if match {
				vCopy := *v
				vCopy.DomainBase = domain.DomainBase{Repo: m}
				results = append(results, &vCopy)
			}
		}

	case *domain.Customer:
		for _, c := range m.Customers {
			match := true
			if vID, ok := conditions["vendor_id = ?"]; ok {
				if c.VendorID != vID {
					match = false
				}
			}
			if nick, ok := conditions["nickname = ?"]; ok {
				if c.Nickname != nick {
					match = false
				}
			}
			if doc, ok := conditions["document = ?"]; ok {
				if c.Document == nil || *c.Document != doc {
					match = false
				}
			}
			if nameLike, ok := conditions["name like ?"]; ok {
				expected := strings.Trim(fmt.Sprint(nameLike), "%")
				if !strings.Contains(strings.ToLower(c.Name), strings.ToLower(expected)) {
					match = false
				}
			}
			if status, ok := conditions["status = ?"]; ok {
				if c.Status == nil || *c.Status != status {
					match = false
				}
			}
			if match {
				cCopy := *c
				cCopy.DomainBase = domain.DomainBase{Repo: m}
				results = append(results, &cCopy)
			}
		}

	case *domain.User:
		for _, u := range m.Users {
			match := true
			// user.go uses either "vendor_id" or "vendor_id = ?"
			if vID, ok := conditions["vendor_id"]; ok {
				if u.VendorID != vID {
					match = false
				}
			}
			if vID, ok := conditions["vendor_id = ?"]; ok {
				if u.VendorID != vID {
					match = false
				}
			}
			if uname, ok := conditions["username"]; ok {
				if u.Username != uname {
					match = false
				}
			}
			if uname, ok := conditions["username = ?"]; ok {
				if u.Username != uname {
					match = false
				}
			}
			if email, ok := conditions["email"]; ok {
				if u.Email == nil || *u.Email != email {
					match = false
				}
			}
			if email, ok := conditions["email = ?"]; ok {
				if u.Email == nil || *u.Email != email {
					match = false
				}
			}
			if match {
				uCopy := *u
				uCopy.DomainBase = domain.DomainBase{Repo: m}
				results = append(results, &uCopy)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported model type in mock Find: %T", model)
	}

	return results, nil
}

type mockPublisher struct {
	PublishedCreated []port.CustomerEventData
	PublishedUpdated []port.CustomerEventData
}

func (m *mockPublisher) PublishCustomerCreated(ctx context.Context, vendor string, data port.CustomerEventData) error {
	m.PublishedCreated = append(m.PublishedCreated, data)
	return nil
}

func (m *mockPublisher) PublishCustomerUpdated(ctx context.Context, vendor string, data port.CustomerEventData) error {
	m.PublishedUpdated = append(m.PublishedUpdated, data)
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}
