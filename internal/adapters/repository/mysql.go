package repository

import (
	"errors"
	"reflect"
	"unicode"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// Begin is a method that starts a transaction
// it returns an object that represents a transaction to be used in others methods
func (r *MySql) Begin() interface{} {
	return r.Db.Begin()
}

// Commit commits the transaction
// it receives a string that represents the transaction name
func (r *MySql) Commit(tx interface{}) error {
	if tx == nil {
		return errors.New(pkg.ErrRepoNilTx)
	}
	stx, ok := tx.(*gorm.DB)
	if !ok {
		return errors.New(pkg.ErrRepoInvalidTX)
	}
	stx = stx.Commit()
	if stx.Error != nil {
		return stx.Error
	}
	return nil
}

// Rollback rolls back the transaction
// it receives a transaction generate by Begin method
func (r *MySql) Rollback(tx interface{}) error {
	if tx == nil {
		return errors.New(pkg.ErrRepoNilTx)
	}
	stx, ok := tx.(*gorm.DB)
	if !ok {
		return errors.New(pkg.ErrRepoInvalidTX)
	}
	stx = stx.Rollback()
	if stx.Error != nil {
		return stx.Error
	}
	return nil
}

// Add adds a object to the database
// it receives the object and the transaction
// transaction have to be started before calling this method
func (r *MySql) Add(tx interface{}, obj interface{}) error {
	stx, err := r.format(tx, obj)
	if err != nil {
		return err
	}
	stx = stx.Session(&gorm.Session{})
	stx.Create(obj)
	if stx.Error != nil {
		return stx.Error
	}
	return nil
}

// Get gets a object from the database by id
// it receives the object, the id and the transaction
// transaction have to be started before calling this method
func (r *MySql) Get(tx interface{}, obj interface{}, id string, lock bool) (bool, error) {
	stx, err := r.format(tx, obj)
	if err != nil {
		return false, err
	}
	stx = stx.Session(&gorm.Session{})
	if lock {
		if stx = stx.Clauses(clause.Locking{Strength: "UPDATE"}); stx.Error != nil {
			return false, stx.Error
		}
	}
	stx = stx.Table(obj.(port.Domain).TableName()).First(obj, "ID = ?", id)
	if stx.Error == nil {
		return true, nil
	}
	if errors.Is(stx.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, stx.Error
}

// Find gets all objects from the database matching the object
// obj represents an object to filter the query and limit is the maximum number of objects to return
// Tx is the transaction name and extras are extra filters commands to the query
// transaction have to be started before calling this method
// The function returns the objects, a boolean indicating if the limit was crossed and an error
// Use -1 to cancel the limit
func (r *MySql) Find(tx interface{}, obj interface{}, limit int, lock bool, extras ...interface{}) (interface{}, bool, error) {
	stx, err := r.format(tx, obj)
	if err != nil {
		return nil, false, err
	}
	result, err := r.find(stx, obj, limit, lock, extras...)
	if err != nil {
		return nil, false, err
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
// it receives the object and the transaction
// transaction have to be started before calling this method
func (r *MySql) Save(tx interface{}, obj interface{}) error {
	stx, err := r.format(tx, obj)
	if err != nil {
		return err
	}
	stx = stx.Session(&gorm.Session{})
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
func (r *MySql) Delete(tx interface{}, obj interface{}, extras ...interface{}) error {
	stx, err := r.format(tx, obj)
	if err != nil {
		return err
	}
	stx = stx.Session(&gorm.Session{})
	stx, err = r.where(stx, reflect.TypeOf(obj).Elem(), obj, extras...)
	if err != nil {
		return err
	}
	stx = stx.Delete(obj)
	if stx.Error != nil {
		return stx.Error
	}
	return nil
}

// find finds object based on obj, limit and extras params
func (r *MySql) find(stx *gorm.DB, obj interface{}, limit int, lock bool, extras ...interface{}) (interface{}, error) {
	stx = stx.Session(&gorm.Session{})
	sob := reflect.TypeOf(obj).Elem()
	var err error
	stx, err = r.where(stx, sob, obj, extras...)
	if err != nil {
		return nil, err
	}
	if limit > 0 {
		stx = stx.Limit(limit + 1)
	}
	result := reflect.New(reflect.SliceOf(sob)).Interface()
	if lock {
		if stx = stx.Clauses(clause.Locking{Strength: "UPDATE"}); stx.Error != nil {
			return nil, err
		}
	}
	if stx = stx.Find(result); stx.Error != nil {
		return nil, stx.Error
	}
	return result, nil
}

// format formats input parameters
func (r *MySql) format(tx interface{}, obj interface{}) (*gorm.DB, error) {
	if tx == nil {
		return nil, errors.New(pkg.ErrRepoNilTx)
	}
	if obj == nil {
		return nil, errors.New(pkg.ErrRepoNilObject)
	}
	stx, ok := tx.(*gorm.DB)
	if !ok {
		return nil, errors.New(pkg.ErrRepoInvalidTX)
	}
	if _, ok := obj.(port.Domain); !ok {
		return nil, errors.New(pkg.ErrRepoInvalidObject)
	}
	return stx, nil
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
