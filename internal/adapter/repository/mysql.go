package repository

import (
	"errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DB_DNS = "MYSQL_INVOICE_DNS"
)

// RepoMySql is the repository handler for the application
type MySql struct {
	Db *gorm.DB
}

// NewRepository creates a new repository handler
func NewRepository(dns string) (*MySql, error) {
	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	return &MySql{Db: db}, nil
}

// Migrate migrates the database
func (r *MySql) Migrate(domain []interface{}) error {
	for _, d := range domain {
		if err := r.Db.AutoMigrate(d); err != nil {
			return err
		}
	}
	return nil
}

// Add adds a object to the database
func (r *MySql) Add(obj interface{}) error {
	return r.Db.Create(obj).Error
}

// Delete deletes a object from the database by id
func (r *MySql) Delete(obj interface{}, id string) error {
	return r.Db.Delete(obj, id).Error
}

// Get gets a object from the database by id
func (r *MySql) Get(obj interface{}, id string) (bool, error) {
	tx := r.Db.First(obj, "ID = ?", id)
	if tx.Error == nil {
		return true, nil
	}
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, tx.Error
}
