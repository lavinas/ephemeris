package pkg

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Command is a struct that represents a command
type Command struct {
	name    string
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

// MarshalSlice is a function that converts a slice of structs to a string
func (c *Commands) Marshal(v interface{}, args ...string) string {
	rvl := c.getInputSlice(v)
	if len(rvl) == 0 {
		return ""
	}
	ret := c.getValuesSlice(rvl, slices.Contains(args, "nokeys"))
	if slices.Contains(args, "trim") {
		ret = c.trimTable(ret)
	}
	return c.mountTable(ret)
}

// UnmarshalOne is a function that converts a string to a struct
func (c *Commands) FindOne(data string, v []interface{}) (interface{}, error) {
	for _, i := range v {
		st := reflect.TypeOf(i).Elem()
		if err := c.checkFields(st); err != nil {
			return nil, err
		}
		tags, err := c.getTags(st, Fieldtag, "keys")
		if err != nil {
			return nil, err
		}
		ss := strings.Split(data, " ")
		if err := c.checkDuplicatedComms(ss, tags); err != nil {
			continue
		}
		c.mapValues(tags, ss)
		if err := c.checkValues(tags); err != nil {
			continue
		}
		return i, nil
	}
	return nil, errors.New(ErrorCommandNotFound)
}

// ToStruc is a function that converts a string to a struct
func (c *Commands) Unmarshal(data string, v interface{}) error {
	ss := strings.Split(data, " ")
	st := reflect.TypeOf(v).Elem()
	if err := c.checkFields(st); err != nil {
		return err
	}
	tags, err := c.getTags(st, Fieldtag)
	if err != nil {
		return err
	}
	if err := c.checkDuplicatedComms(ss, tags); err != nil {
		return err
	}
	c.mapValues(tags, ss)
	if err := c.checkValues(tags); err != nil {
		return err
	}
	c.setFields(v, tags)
	return nil
}

// getInputSlice is a function that returns a slice of reflect.Values
func (c *Commands) getInputSlice(v interface{}) []reflect.Value {
	vl := reflect.ValueOf(v)
	if vl.Kind() == reflect.Ptr {
		vl = vl.Elem()
	}
	rvl := []reflect.Value{}
	if vl.Kind() != reflect.Slice {
		rvl = append(rvl, vl)
	} else {
		for i := 0; i < vl.Len(); i++ {
			rvl = append(rvl, vl.Index(i))
		}
	}
	return rvl
}

// getValuesSlice is a function that returns a slice of strings
func (c *Commands) getValuesSlice(values []reflect.Value, nokeys bool) [][]string {
	ret := make([][]string, len(values)+1)
	for i := 0; i < values[0].NumField(); i++ {
		if nokeys {
			tag := c.getTag(values[0].Type().Field(i), Fieldtag)
			if tag == nil || tag.iskey {
				continue
			}
		}
		ret[0] = append(ret[0], values[0].Type().Field(i).Name)
	}
	for i := 0; i < len(values); i++ {
		for j := 0; j < values[i].NumField(); j++ {
			if nokeys {
				tag := c.getTag(values[i].Type().Field(j), Fieldtag)
				if tag == nil || tag.iskey {
					continue
				}
			}
			ret[i+1] = append(ret[i+1], values[i].Field(j).String())
		}
	}
	return ret
}

// checkDuplicatedWords is a function that checks if all command words are unique
func (c *Commands) checkDuplicatedComms(ss []string, tags map[string]*Command) error {
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
func (c *Commands) checkFields(st reflect.Type) error {
	for i := 0; i < st.NumField(); i++ {
		if st.Field(i).Type.String() != "string" {
			return errors.New(ErrorNotStringField)
		}
	}
	return nil
}

// getTags is a function that returns all tags of a struct
func (c *Commands) getTags(st reflect.Type, tagname string, args ...string) (map[string]*Command, error) {
	ret := map[string]*Command{}
	keys := slices.Contains(args, "keys")
	for i := 0; i < st.NumField(); i++ {
		tag := c.getTag(st.Field(i), tagname)
		if tag == nil {
			continue
		}
		if _, ok := ret[tag.name]; ok {
			return nil, errors.New(ErrorStructDuplicated)
		}
		if keys && !tag.iskey {
			continue
		}
		ret[tag.name] = tag
	}
	return ret, nil
}

// getTag is a function that returns a tag struct of a struct
func (c *Commands) getTag(field reflect.StructField, tagname string) *Command {
	tag := field.Tag.Get(tagname)
	if tag == "" {
		return nil
	}
	name, notnull, iskey := c.splitValues(tag)
	if name != "" {
		return &Command{name: name, field: field.Name, iskey: iskey, notnull: notnull}
	}
	return nil
}

// splitValues is a function that splits the values of a tag into name, notnull and iskey
func (c *Commands) splitValues(tag string) (string, bool, bool) {
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
func (c *Commands) mapValues(tags map[string]*Command, ss []string) {
	for tag, field := range tags {
		if !slices.Contains(ss, tag) {
			field.isfound = "false"
			continue
		} else {
			field.isfound = "true"
		}
		param := c.getValue(tag, tags, ss)
		field.value = strings.TrimSpace(param)
	}
}

// getValue is a function that returns the value of a tag
func (c *Commands) getValue(tag string, tags map[string]*Command, ss []string) string {
	value := ""
	for j := slices.Index(ss, tag) + 1; j < len(ss); j++ {
		if _, ok := tags[ss[j]]; ok {
			break
		}
		if ss[j][0] == '.' {
			ss[j] = ss[j][1:]
		}
		if ss[j] != "" {
			value += ss[j] + " "
		}
	}
	return value
}

// checkValues is a function that checks if all values are correct
func (c *Commands) checkValues(tags map[string]*Command) error {
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
func (c *Commands) setFields(v interface{}, tags map[string]*Command) {
	for _, i := range tags {
		field := reflect.ValueOf(v).Elem().FieldByName(i.field)
		field.SetString(i.value)
	}
}

// trimTable is a function that trims a table
func (c *Commands) trimTable(table [][]string) [][]string {
	fillCols := map[int]bool{}
	fillLines := map[int]bool{}
	for i := 1; i < len(table); i++ {
		for j := 0; j < len(table[i]); j++ {
			if table[i][j] != "" {
				fillCols[j] = true
				fillLines[i] = true
			}
		}
	}
	ret := make([][]string, 0)
	line := []string{}
	for i := 0; i < len(table[0]); i++ {
		if fillCols[i] {
			line = append(line, table[0][i])
		}
	}
	if len(line) > 0 {
		ret = append(ret, line)
	}
	for i := 1; i < len(table); i++ {
		if !fillLines[i] {
			continue
		}
		line := []string{}
		for j := 0; j < len(table[i]); j++ {
			if fillCols[j] {
				line = append(line, table[i][j])
			}
		}
		if len(line) > 0 {
			ret = append(ret, line)
		}
	}
	return ret
}


// mountTable is a function that mounts a table
func (c *Commands) mountTable(table [][]string) string {
	ret := ""
	// get number of columns from the first table row
	columnLengths := make([]int, len(table[0]))
	for _, line := range table {
		for i, val := range line {
			if len(val) > columnLengths[i] {
				columnLengths[i] = len(val)
			}
		}
	}
	var lineLength int
	for _, c := range columnLengths {
		lineLength += c + 3
	}
	lineLength += 1
	for i, line := range table {
		if i == 0 {
			ret += fmt.Sprintf("+%s+\n", strings.Repeat("-", lineLength-2))
		}
		for j, val := range line {
			ret += fmt.Sprintf("| %-*s ", columnLengths[j], val)
			if j == len(line)-1 {
				ret += "|\n"
			}
		}
		if i == 0 || i == len(table)-1 {
			ret += fmt.Sprintf("+%s+\n", strings.Repeat("-", lineLength-2))
		}
	}
	return ret
}
