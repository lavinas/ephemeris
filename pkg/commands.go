package pkg

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	ErrorNotStringField = "not all fields are strings"
	ErrorNotNullField   = "field %s is empty"
	fieldtag            = "command"
	tagname             = "name"
	tagnotnull          = "not null"
)

type Command struct {
	field   string
	notnull bool
	value   string
}

// Texts is a struct that groups all texts functionalities
type Commands struct {
}

// NewStrings is a function that returns a new Strings
func NewCommands() *Commands {
	return &Commands{}
}

// ToStruc is a function that converts a string to a struct
func (s *Commands) Unmarshal(data string, v interface{}) error {
	ss := strings.Split(data, " ")
	st := reflect.TypeOf(v).Elem()
	if err := s.checkFieldsType(st); err != nil {
		return err
	}
	tags := s.getTags(st, fieldtag)
	if err := s.mapValues(tags, ss); err != nil {
		return err
	}
	s.setFields(v, tags)
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
func (s *Commands) getTags(st reflect.Type, tag string) map[string]*Command {
	ret := map[string]*Command{}
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag.Get(tag)
		if tag == "" {
			continue
		}
		name := ""
		notnull := false
		fds := strings.Split(tag, ";")
		for _, fd := range fds {
			if strings.Contains(fd, tagname) {
				name = strings.Split(fd, ":")[1]
			}
			if strings.Contains(fd, tagnotnull) {
				notnull = true
			}
		}
		if name == "" {
			ret[st.Field(i).Name] = &Command{field: st.Field(i).Name, notnull: notnull, value: ""}
		}
	}
	return ret
}

// mapValues is a function that maps values to a struct
func (s *Commands) mapValues(tags map[string]*Command, ss []string) error {
	for tag, field := range tags {
		if !slices.Contains(ss, tag) {
			continue
		}
		param := ""
		for j := slices.Index(ss, tag) + 1; j < len(ss); j++ {
			if _, ok := tags[ss[j]]; !ok {
				break
			}
			param += ss[j] + " "
		}
		field.value = strings.TrimSpace(param)
		if field.notnull && field.value == "" {
			return fmt.Errorf(ErrorNotNullField, tag)
		}
	}
	return nil
}

// setFields is a function that sets the fields of a DTO
func (s *Commands) setFields(v interface{}, tags map[string]*Command) {
	for _, i := range tags {
		field := reflect.ValueOf(v).Elem().FieldByName(i.field)
		field.SetString(i.value)
	}
}
