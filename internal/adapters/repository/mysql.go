package repository

import (
	"errors"
	"reflect"
	"time"
	"unicode"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DB_DNS      = "MYSQL_INVOICE_DNS"
	ErrNoFilter = "no fields where provided on base object"
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
	tx := r.Db.Begin()
	tx = tx.Create(obj)
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}
	tx.Commit()
	return nil
}

// Delete deletes a object from the database by id
func (r *MySql) Delete(obj interface{}, id string) error {
	tx := r.Db.Begin()
	tx = tx.Delete(obj, "ID = ?", id)
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}
	tx.Commit()
	return nil
}

// Get gets a object from the database by id
func (r *MySql) Get(obj interface{}, id string) (bool, error) {
	tx := r.Db.Begin()
	defer tx.Rollback()
	tx = tx.First(obj, "ID = ?", id)
	if tx.Error == nil {
		return true, nil
	}
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, tx.Error
}

// Save saves a object to the database
func (r *MySql) Save(obj interface{}) error {
	tx := r.Db.Begin()
	tx = tx.Save(obj)
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}
	tx.Commit()
	return nil
}

// Find gets all objects from the database matching the object
func (r *MySql) Find(base interface{}) (interface{}, error) {
	sob := reflect.TypeOf(base).Elem()
	result := reflect.New(reflect.SliceOf(sob)).Interface()
	tx := r.Db.Begin()
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