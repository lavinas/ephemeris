package pkg

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	ErrorTagNameNotFound = "tag name not found"
	ErrorNotStringField  = "not all fields are strings"
	ErrorKeyNotFound     = "tag %s not found"
	ErrorNotNullField    = "tag %s is null"
	fieldtag             = "command"
	tagname              = "name:"
	tagnotnull           = "not null"
	tagkey               = "key"
)

// Command is a struct that represents a command
type Command struct {
	field   string
	iskey   bool
	notnull bool
	isfound string
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
	s.mapValues(tags, ss)
	if err := s.checkValues(tags); err != nil {
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
		name, notnull, iskey := s.splitValues(tag)
		if name != "" {
			ret[name] = &Command{field: st.Field(i).Name, iskey: iskey, notnull: notnull}
		}

	}
	return ret
}

// splitValues is a function that splits the values of a tag into name, notnull and iskey
func (s *Commands) splitValues(tag string) (string, bool, bool) {
	fields := strings.Split(tag, ";")
	name := ""
	notnull := false
	iskey := false
	for _, fd := range fields {
		if strings.Contains(fd, tagname) {
			name = strings.Split(fd, ":")[1]
		}
		if strings.Contains(fd, tagnotnull) {
			notnull = true
		}
		if strings.Contains(fd, tagkey) {
			iskey = true
		}
	}
	return name, notnull, iskey
}

// mapValues is a function that maps values to a struct
func (s *Commands) mapValues(tags map[string]*Command, ss []string) {
	for tag, field := range tags {
		if !slices.Contains(ss, tag) {
			field.isfound = "false"
			continue
		} else {
			field.isfound = "true"
		}
		param := s.getValue(tag, tags, ss)
		field.value = strings.TrimSpace(param)
	}
}

// getValue is a function that returns the value of a tag
func (s *Commands) getValue(tag string, tags map[string]*Command, ss []string) string {
	param := ""
	for j := slices.Index(ss, tag) + 1; j < len(ss); j++ {
		if _, ok := tags[ss[j]]; ok {
			break
		}
		param += ss[j] + " "
	}
	return param
}

// checkValues is a function that checks if all values are correct
func (s *Commands) checkValues(tags map[string]*Command) error {
	message := ""
	for tag, field := range tags {
		if field.isfound == "false" && (field.iskey || field.notnull){
			message += fmt.Sprintf(ErrorKeyNotFound, tag) + " | "
		} else if field.notnull && field.value == "" {
			message += fmt.Sprintf(ErrorNotNullField, tag) + " | "
		}
	}
	if message != "" {
		return errors.New(message[:len(message)-3])
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
