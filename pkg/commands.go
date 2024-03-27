package pkg

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	ErrorStructDuplicated  = "command words are duplicated in the struct"
	ErrorCommandDuplicated = "more than one command found. Try use . in front of the parameter words if parameter has command words"
	ErrorWordDuplicated    = "command words are duplicated %s. Try use . in front of the parameter words if parameter has command words"
	ErrorTagNameNotFound   = "tag name not found"
	ErrorNotStringField    = "not all fields are strings"
	ErrorKeyNotFound       = "tag %s not found"
	ErrorNotNullField      = "tag %s is null"
	Fieldtag               = "command"
	Tagname                = "name:"
	Tagnotnull             = "not null"
	Tagkey                 = "key"
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

// Marshal is a function that converts a struct to a string
func (s *Commands) Marshal(v interface{}) string {
	st := reflect.TypeOf(v).Elem()
	ret := ""
	for i := 0; i < st.NumField(); i++ {
		ret += fmt.Sprintf("%s: %s | ", st.Field(i).Name, reflect.ValueOf(v).Elem().Field(i).String())
	}
	return ret[:len(ret)-3]
}

// Choose is a function that chooses the correct struct to return
func (s *Commands) UnmarshalOne(data string, v []interface{}) (interface{}, error) {
	found := []interface{}{}
	for _, i := range v {
		if err := s.Unmarshal(data, i); err != nil {
			continue
		}
		found = append(found, i)
	}
	if len(found) == 0 {
		return nil, errors.New(ErrorTagNameNotFound)
	}
	if len(found) > 1 {
		return nil, errors.New(ErrorCommandDuplicated)
	}
	return found[0], nil
}

// ToStruc is a function that converts a string to a struct
func (s *Commands) Unmarshal(data string, v interface{}) error {
	ss := strings.Split(data, " ")
	st := reflect.TypeOf(v).Elem()
	if err := s.checkFields(st); err != nil {
		return err
	}
	tags, err := s.getTags(st, Fieldtag)
	if err != nil {
		return err
	}
	if err := s.checkDuplicatedComms(ss, tags); err != nil {
		return err
	}
	s.mapValues(tags, ss)
	if err := s.checkValues(tags); err != nil {
		return err
	}
	s.setFields(v, tags)
	return nil
}

// checkDuplicatedWords is a function that checks if all command words are unique
func (s *Commands) checkDuplicatedComms(ss []string, tags map[string]*Command) error {
	wordMap := map[string]int{}
	for _, i := range ss {
		if tags[i] != nil {
			wordMap[i]++
		}
	}
	errorStr := ""
	for k, v := range wordMap {
		if v > 1 {
			errorStr += k + ", "
		}
	}
	if errorStr != "" {
		return fmt.Errorf(ErrorWordDuplicated, errorStr[:len(errorStr)-2])
	}
	return nil
}

// checkFieldsType is a function that checks if all fields of a struct are strings
func (s *Commands) checkFields(st reflect.Type) error {
	for i := 0; i < st.NumField(); i++ {
		if st.Field(i).Type.String() != "string" {
			return errors.New(ErrorNotStringField)
		}
	}
	return nil
}

// getTags is a function that returns all tags of a struct
func (s *Commands) getTags(st reflect.Type, tag string) (map[string]*Command, error) {
	ret := map[string]*Command{}
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag.Get(tag)
		if tag == "" {
			continue
		}
		name, notnull, iskey := s.splitValues(tag)
		if name != "" {
			if _, ok := ret[name]; ok {
				return nil, errors.New(ErrorStructDuplicated)
			}

			ret[name] = &Command{field: st.Field(i).Name, iskey: iskey, notnull: notnull}
		}
	}
	return ret, nil
}

// splitValues is a function that splits the values of a tag into name, notnull and iskey
func (s *Commands) splitValues(tag string) (string, bool, bool) {
	fields := strings.Split(tag, ";")
	name := ""
	notnull := false
	iskey := false
	for _, fd := range fields {
		if strings.Contains(fd, Tagname) {
			name = strings.Split(fd, ":")[1]
		}
		if strings.Contains(fd, Tagnotnull) {
			notnull = true
		}
		if strings.Contains(fd, Tagkey) {
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
		if ss[j][0] == '.' {
			param += ss[j][1:]
		}
		if ss[j] != "" {
			param += ss[j] + " "
		}
	}
	return param
}

// checkValues is a function that checks if all values are correct
func (s *Commands) checkValues(tags map[string]*Command) error {
	message := ""
	for tag, field := range tags {
		if field.isfound == "false" && (field.iskey || field.notnull) {
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
