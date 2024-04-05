package repository

import (
	"errors"
	"reflect"
	"time"
	"unicode"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/lavinas/ephemeris/internal/port"
)

const (
	DB_DNS      = "MYSQL_INVOICE_DNS"
	ErrNoFilter = "no fields where provided on base object"
)

// RepoMySql is the repository handler for the application
type MySql struct {
	Db *gorm.DB
	tx *gorm.DB
}

// NewRepository creates a new repository handler
func NewRepository(dns string) (*MySql, error) {
	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	return &MySql{Db: db}, nil
}

// Begin starts a transaction
func (r *MySql) Begin() error {
	if r.tx != nil {
		return errors.New("transaction already started")
	}
	r.tx = r.Db.Begin()
	return nil
}

// Commit commits the transaction
func (r *MySql) Commit() error {
	if r.tx.Error == nil {
		return errors.New("no transaction to commit")
	}
	r.tx.Commit()
	r.tx = nil
	return nil
}

// Rollback rolls back the transaction
func (r *MySql) Rollback() error {
	if r.tx.Error != nil {
		return errors.New("no transaction to rollback")
	}
	r.tx.Rollback()
	r.tx = nil
	return nil
}

// Close closes the database connection
func (r *MySql) Close() {
	db, err := r.Db.DB()
	if err != nil {
		return
	}
	db.Close()
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
	fmt.Println(0, obj)
	r.tx = r.tx.Create(obj)
	if r.tx.Error != nil {
		fmt.Println(1000, r.tx.Error)
		return r.tx.Error
	}
	return nil
}

// Delete deletes a object from the database by id
func (r *MySql) Delete(obj interface{}, id string) error {
	r.tx = r.tx.Delete(obj, "ID = ?", id)
	if r.tx.Error != nil {
		return r.tx.Error
	}
	return nil
}

// Get gets a object from the database by id
func (r *MySql) Get(obj interface{}, id string) (bool, error) {
	d := obj.(port.Domain)
	name := d.TableName()
	r.tx = r.tx.Table(name).First(obj, "ID = ?", id)
	if r.tx.Error == nil {
		return true, nil
	}
	if errors.Is(r.tx.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	fmt.Println(2, r.tx.Error)
	return false, r.tx.Error
}

// Save saves a object to the database
func (r *MySql) Save(obj interface{}) error {
	r.tx = r.tx.Save(obj)
	if r.tx.Error != nil {
		return r.tx.Error
	}
	return nil
}

// Find gets all objects from the database matching the object
func (r *MySql) Find(base interface{}) (interface{}, error) {
	sob := reflect.TypeOf(base).Elem()
	result := reflect.New(reflect.SliceOf(sob)).Interface()
	tx := r.tx.Begin()
	defer tx.Rollback()
	tx, err := r.where(tx, sob, base)
	if err != nil {
		return nil, err
	}
	if tx = tx.Find(result); tx.Error != nil {
		return nil, tx.Error
	}
	if reflect.ValueOf(result).Elem().Len() == 0 {
		return nil, nil
	}
	return result, nil
}

// where is a method that filters the query
func (r *MySql) where(tx *gorm.DB, sob reflect.Type, base interface{}) (*gorm.DB, error) {
	filtered := false
	for i := 0; i < sob.NumField(); i++ {
		if r.isEmpty(reflect.ValueOf(base).Elem().Field(i)) {
			continue
		}
		filtered = true
		fName := r.fieldName(sob.Field(i).Name)
		tx = tx.Where(fName+" = ?", reflect.ValueOf(base).Elem().Field(i).Interface())
	}
	if !filtered {
		return nil, errors.New(ErrNoFilter)
	}
	return tx, nil
}


// isEmpty is a method that returns true if the value is empty
func (r *MySql) isEmpty(value reflect.Value) bool {
	typ := value.Type()
	val := value.Interface()
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return true
	}
	if typ == reflect.TypeOf(time.Time{}) && val.(time.Time).IsZero() {
		return true
	}
	if typ == reflect.TypeOf("") && val == "" {
		return true
	}
	if typ == reflect.TypeOf(0) && val == 0 {
		return true
	}
	return false
}

// fieldName is a method that returns the field name
func (r *MySql) fieldName (field string) string {
	ret := ""
	isLower := false 
	for _, ch := range field {
		if unicode.IsUpper(ch) && isLower {
			ret += "_"
		}
		isLower = unicode.IsLower(ch)
		ret += string(unicode.ToLower(ch))
	}
	return ret
}