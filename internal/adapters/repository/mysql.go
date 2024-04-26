package repository

import (
	"errors"
	"reflect"
	"unicode"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
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

// Begin starts a transaction
func (r *MySql) Begin() error {
	r.Db = r.Db.Begin()
	if r.Db.Error != nil {
		return r.Db.Error
	}
	return nil
}

// Commit commits the transaction
func (r *MySql) Commit() error {
	r.Db = r.Db.Commit()
	if r.Db.Error != nil {
		return r.Db.Error
	}
	return nil
}

// Rollback rolls back the transaction
func (r *MySql) Rollback() error {
	r.Db = r.Db.Rollback()
	if r.Db.Error != nil {
		return r.Db.Error
	}
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
	tx := r.Db.Session(&gorm.Session{})
	tx.Create(obj)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// Delete deletes a object from the database by id
func (r *MySql) Delete(obj interface{}, id string) error {
	tx := r.Db.Session(&gorm.Session{})
	tx = tx.Delete(obj, "ID = ?", id)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// Get gets a object from the database by id
func (r *MySql) Get(obj interface{}, id string) (bool, error) {
	d := obj.(port.Domain)
	name := d.TableName()
	tx := r.Db.Session(&gorm.Session{})
	tx = tx.Table(name).First(obj, "ID = ?", id)
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
	tx := r.Db.Session(&gorm.Session{})
	tx = tx.Save(obj)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// Find gets all objects from the database matching the object
// Base represents a base object to filter the query and limit is the maximum number of objects to return
// The function returns the objects, a boolean indicating if the limit was crossed and an error
// Use -1 to cancel the limit
func (r *MySql) Find(base interface{}, limit int) (interface{}, bool, error) {
	sob := reflect.TypeOf(base).Elem()
	result := reflect.New(reflect.SliceOf(sob)).Interface()
	tx := r.Db.Session(&gorm.Session{})
	var err error
	tx, err = r.where(tx, sob, base)
	if err != nil {
		return nil, false, err
	}
	if tx = tx.Limit(limit + 1).Find(result); tx.Error != nil {
		return nil, false, tx.Error
	}
	if reflect.ValueOf(result).Elem().Len() == 0 {
		return nil, false, nil
	}
	crossLimit := false
	if limit != -1 && reflect.ValueOf(result).Elem().Len() > limit {
		reflect.ValueOf(result).Elem().SetLen(limit)
		crossLimit = true
	}
	return result, crossLimit, nil
}

// where is a method that filters the query
func (r *MySql) where(tx *gorm.DB, sob reflect.Type, base interface{}) (*gorm.DB, error) {
	for i := 0; i < sob.NumField(); i++ {
		isgorm := sob.Field(i).Tag.Get("gorm")
		if isgorm == "-" || isgorm == ""{
			continue
		}
		if sob.Field(i).Type.Kind() == reflect.Struct {
			etype := reflect.TypeOf(base).Elem().Field(i).Type
			eval := reflect.ValueOf(base).Elem().Field(i).Interface()
			var err error
			if tx, err = r.where(tx, etype, eval); err != nil {
				return nil, err
			}
			continue
		}
		elem := reflect.ValueOf(base).Elem().Field(i).Interface()
		if pkg.IsEmpty(elem) {
			continue
		}
		fName := r.fieldName(sob.Field(i).Name)
		tx = tx.Where(fName+" = ?", elem)
		if i == 0 {
			tx = tx.Session(&gorm.Session{})
		}
	}
	return tx, nil
}

// fieldName is a method that returns the field name
func (r *MySql) fieldName(field string) string {
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
