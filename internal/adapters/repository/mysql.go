package repository

import (
	"errors"
	"reflect"
	"unicode"

	"github.com/google/uuid"
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
	Tx map[string]*gorm.DB
}

// NewRepository creates a new repository handler
func NewRepository(dns string) (*MySql, error) {
	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	tx := make(map[string]*gorm.DB)
	return &MySql{Db: db, Tx: tx}, nil
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
// it receives a slice of interfaces that represents the domain
func (r *MySql) Migrate(domain []interface{}) error {
	for _, d := range domain {
		if err := r.Db.AutoMigrate(d); err != nil {
			return err
		}
	}
	return nil
}

// NewTransaction creates a new transaction
func (r *MySql) NewTransaction() string {
	return uuid.New().String()
}

// Begin is a method that starts a transaction
// it receives a string that represents the transaction name
func (r *MySql) Begin(tx string) error {
	if _, ok := r.Tx[tx]; ok {
		return errors.New(pkg.ErrRepoTransactionStarted)
	}
	r.Tx[tx] = r.Db.Begin()
	return nil
}

// Commit commits the transaction
// it receives a string that represents the transaction name
func (r *MySql) Commit(tx string) error {
	if _, ok := r.Tx[tx]; !ok {
		return errors.New(pkg.ErrRepoTransactionNotStarted)
	}
	r.Tx[tx] = r.Tx[tx].Commit()
	if r.Tx[tx].Error != nil {
		return r.Tx[tx].Error
	}
	delete(r.Tx, tx)
	return nil
}

// Rollback rolls back the transaction
// it receives a string that represents the transaction name
func (r *MySql) Rollback(tx string) error {
	if _, ok := r.Tx[tx]; !ok {
		return errors.New(pkg.ErrRepoTransactionNotStarted)
	}
	r.Tx[tx] = r.Tx[tx].Rollback()
	if r.Tx[tx].Error != nil {
		return r.Tx[tx].Error
	}
	delete(r.Tx, tx)
	return nil
}

// Add adds a object to the database
// it receives the object and the transaction name
// transaction have to be started before calling this method
func (r *MySql) Add(obj interface{}, tx string) error {
	if _, ok := r.Tx[tx]; !ok {
		return errors.New(pkg.ErrRepoTransactionNotStarted)
	}
	stx := r.Tx[tx].Session(&gorm.Session{})
	stx.Create(obj)
	if stx.Error != nil {
		return stx.Error
	}
	return nil
}

// Get gets a object from the database by id
// it receives the object, the id and the transaction name
// transaction have to be started before calling this method
func (r *MySql) Get(obj interface{}, id string, tx string) (bool, error) {
	if _, ok := r.Tx[tx]; !ok {
		return false, errors.New(pkg.ErrRepoTransactionNotStarted)
	}
	d := obj.(port.Domain)
	name := d.TableName()
	stx := r.Tx[tx].Session(&gorm.Session{})
	stx = stx.Table(name).First(obj, "ID = ?", id)
	if stx.Error == nil {
		return true, nil
	}
	if errors.Is(stx.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, stx.Error
}

// Find gets all objects from the database matching the object
// Base represents a base object to filter the query and limit is the maximum number of objects to return
// Tx is the transaction name and extras are extra filters commands to the query
// transaction have to be started before calling this method
// The function returns the objects, a boolean indicating if the limit was crossed and an error
// Use -1 to cancel the limit
func (r *MySql) Find(base interface{}, limit int, tx string, extras ...interface{}) (interface{}, bool, error) {
	if _, ok := r.Tx[tx]; !ok {
		return nil, false, errors.New(pkg.ErrRepoTransactionNotStarted)
	}
	sob := reflect.TypeOf(base).Elem()
	result := reflect.New(reflect.SliceOf(sob)).Interface()
	stx := r.Tx[tx].Session(&gorm.Session{})
	var err error
	stx, err = r.where(stx, sob, base, extras...)
	if err != nil {
		return nil, false, err
	}
	if limit > 0 {
		stx = stx.Limit(limit + 1)
	}
	if stx = stx.Find(result); stx.Error != nil {
		return nil, false, stx.Error
	}
	if reflect.ValueOf(result).Elem().Len() == 0 {
		return nil, false, nil
	}
	crossLimit := false
	if limit > 0 && reflect.ValueOf(result).Elem().Len() > limit {
		reflect.ValueOf(result).Elem().SetLen(limit)
		crossLimit = true
	}
	return result, crossLimit, nil
}

// Save saves a object to the database
// it receives the object and the transaction name
// transaction have to be started before calling this method
func (r *MySql) Save(obj interface{}, tx string) error {
	if _, ok := r.Tx[tx]; !ok {
		return errors.New(pkg.ErrRepoTransactionNotStarted)
	}
	stx := r.Tx[tx].Session(&gorm.Session{})
	stx = stx.Save(obj)
	if stx.Error != nil {
		return stx.Error
	}
	return nil
}

// Delete deletes a object from the database by id
// it receives the object, the id and the transaction name
// Tx is the transaction name and extras are extra filters commands to the query
// transaction have to be started before calling this method
func (r *MySql) Delete(obj interface{}, tx string, extras ...interface{}) error {
	if _, ok := r.Tx[tx]; !ok {
		return errors.New(pkg.ErrRepoTransactionNotStarted)
	}
	stx := r.Tx[tx].Session(&gorm.Session{})
	stx, err := r.where(stx, reflect.TypeOf(obj).Elem(), obj, extras...)
	if err != nil {
		return err
	}
	stx = stx.Delete(obj)
	if stx.Error != nil {
		return stx.Error
	}
	return nil
}

// where is a method that filters the query
func (r *MySql) where(tx *gorm.DB, sob reflect.Type, base interface{}, extras ...interface{}) (*gorm.DB, error) {
	if sob.Kind() == reflect.Ptr {
		sob = sob.Elem()
	}
	for i := 0; i < sob.NumField(); i++ {
		isgorm := sob.Field(i).Tag.Get("gorm")
		if isgorm == "-" || isgorm == "" {
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
	for _, extra := range extras {
		tx = tx.Where(extra)
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
