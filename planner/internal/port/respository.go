package port

import (
	"planner/internal/domain"
)

// Repository defines the interface for interacting with the data storage layer.
type Repository interface {
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Save(model interface{}) error
	Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error)
	FindCount(conditions map[string]interface{}) (int64, error)
	FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error)
	Close() error
	FindCustomers(page, pageSize int, vendorID int64, name, nickname,
		document *string, status *int, email, whatsapp *string) ([]domain.Customer, error)
	GetCustomer(vendorID int64, nickname string) (*domain.Customer, error)
	FindVendors(page, pageSize int, legalName, nickname, document *string,
		accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error)
	GetVendor(nickname string) (*domain.Vendor, error)
}
