package pkg

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"fmt"
)

const (
	ErrorNotStringField = "not all fields are strings"
)

// Texts is a struct that groups all texts functionalities
type Commands struct {
}

// NewStrings is a function that returns a new Strings
func NewCommands() *Commands {
	return &Commands{}
}

// ToStruc is a function that converts a string to a struct
func (s *Commands) Unmarshal(data string, v interface{}, tag string) error {
	ss := strings.Split(data, " ")
	fmt.Println("opa", len(ss), ss)
	st := reflect.TypeOf(v).Elem()
	if err := s.checkFieldsType(st); err != nil {
		return err
	}
	tags := s.getTags(st, tag)
	values := s.mapValues(tags, ss)
	s.setFields(v, values)
	return nil

}

// checkFieldsType is a function that checks if all fields of a struct are strings
func (s *Commands) checkFieldsType(st reflect.Type) error {
	for i := 0; i < st.NumField(); i++ {
		if st.Field(i).Type.String() != "string" {
			return errors.New(ErrorNotStringField)
		}
	}
	return nil
}

// getTags is a function that returns all tags of a struct
func (s *Commands) getTags(st reflect.Type, tag string) map[string]string {
	tags := map[string]string{}
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag.Get(tag)
		if tag == "" {
			continue
		}
		tags[tag] = st.Field(i).Name
	}
	return tags
}

// mapValues is a function that maps values to a struct
func (s *Commands) mapValues(tags map[string]string, ss []string) map[string]string {
	valueMap := map[string]string{}
	for tag, field := range tags {
		if !slices.Contains(ss, tag) {
			continue
		}
		param := ""
		for j := slices.Index(ss, tag) + 1; j < len(ss); j++ {
			if tags[ss[j]] != "" {
				break
			}
			param += ss[j] + " "
		}
		valueMap[field] = strings.TrimSpace(param)
	}
	return valueMap
}

// setFields is a function that sets the fields of a DTO
func (s *Commands) setFields(v interface{}, values map[string]string) {
	for k, i := range values {
		field := reflect.ValueOf(v).Elem().FieldByName(k)
		field.SetString(i)
	}
}
