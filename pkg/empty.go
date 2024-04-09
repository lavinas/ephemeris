package pkg

import (
	// "fmt"
	"time"

	"reflect"
)

var (
	// EmptyMap is a map with no elements
	EmptyMap = map[interface{}]interface{}{
		reflect.TypeOf(int64(0)):    int64(-9223372036854775808),
		reflect.TypeOf(int32(0)):    int32(-2147483648),
		reflect.TypeOf(int16(0)):    int16(-32768),
		reflect.TypeOf(int8(0)):     int8(-128),
		reflect.TypeOf(int(0)):      int(-2147483648),
		reflect.TypeOf(uint64(0)):   uint64(0),
		reflect.TypeOf(uint32(0)):   uint32(0),
		reflect.TypeOf(uint16(0)):   uint16(0),
		reflect.TypeOf(uint8(0)):    uint8(0),
		reflect.TypeOf(uint(0)):     uint(0),
		reflect.TypeOf(float64(0)):  float64(-1.7e+308),
		reflect.TypeOf(float32(0)):  float32(-3.4e+38),
		reflect.TypeOf(time.Time{}): time.Time{},
		reflect.TypeOf(""):          "",
	}
)

// IsEmpty returns true if the interface is empty
func IsEmpty(i interface{}) bool {
	if i == nil {
		return true
	}
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	if t.Kind() == reflect.Ptr {
		if v.IsNil() { 
			return true
		}
		t = t.Elem()
		v = v.Elem()
	}
	if u, ok := EmptyMap[t]; ok {
		if v.Interface() == u {
			return true
		}
	}
	return false
}

// GetEmpty
func GetEmpty(i interface{}) interface{} {
	v := reflect.TypeOf(i)
	if u, ok := EmptyMap[v]; ok {
		return u
	}
	return nil
}
