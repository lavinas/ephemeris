package pkg

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

)

// Command is a struct that represents a command
type Param struct {
	names    []string
	name     string
	field   string
	ftype   string
	iskey   bool
	pos     string
	notnull bool
	value   string
}

// Texts is a struct that groups all texts functionalities
type Commands2 struct {
}

// NewStrings is a function that returns a new Strings
func NewCommands2() *Commands2 {
	return &Commands2{}
}

// MarshalSlice is a function that converts a slice of structs to a string
func (c *Commands2) Marshal(v interface{}, args ...string) string {
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
func (c *Commands2) FindOne(data string, v []interface{}) (interface{}, error) {
	ret := []interface{}{}
	for _, i := range v {
		st := reflect.TypeOf(i).Elem()
		params, err := c.getParams(st, Fieldtag, "keys")
		if err != nil {
			return nil, err
		}
		names, err := c.validateFields(params)
		if err != nil {
			return nil, err
		}
		if err := c.mapValues(data, params, names); err != nil {
			continue
		}
		ret = append(ret, i)
	}
	if len(ret) == 0 {
		return nil, errors.New(ErrorCommandNotFound)
	}
	if len(ret) > 1 {
		return nil, errors.New(ErrorCommandDuplicated)
	}
	return ret[0], nil
}

// ToStruc is a function that converts a string to a struct
func (c *Commands2) Unmarshal(data string, v interface{}) error {
	st := reflect.TypeOf(v).Elem()
	tags, err := c.getParams(st, Fieldtag)
	if err != nil {
		return err
	}
	names, err := c.validateFields(tags) 
	if err != nil {
		return err
	}
	if err := c.mapValues(data, tags, names); err != nil {
		return err
	}
	c.setFields(v, tags)
	return nil
}

// getInputSlice is a function that returns a slice of reflect.Values
func (c *Commands2) getInputSlice(v interface{}) []reflect.Value {
	vl := reflect.ValueOf(v)
	for vl.Kind() == reflect.Ptr || vl.Kind() == reflect.Interface {
		vl = vl.Elem()
	}
	rvl := []reflect.Value{}
	if vl.Kind() != reflect.Slice {
		for vl.Kind() == reflect.Ptr || vl.Kind() == reflect.Interface {
			vl = vl.Elem()
		}
		rvl = append(rvl, vl)
	} else {
		for i := 0; i < vl.Len(); i++ {
			obj := vl.Index(i)
			for obj.Kind() == reflect.Ptr || obj.Kind() == reflect.Interface {
				obj = obj.Elem()
			}
			rvl = append(rvl, obj)
		}
	}
	return rvl
}

// getValuesSlice is a function that returns a slice of strings
func (c *Commands2) getValuesSlice(values []reflect.Value, nokeys bool) [][]string {
	ret := make([][]string, len(values)+1)
	for i := 0; i < values[0].NumField(); i++ {
		if nokeys {
			tag := c.getParam(values[0].Type().Field(i), Fieldtag)
			if tag == nil || tag.iskey {
				continue
			}
		}
		ret[0] = append(ret[0], values[0].Type().Field(i).Name)
	}
	for i := 0; i < len(values); i++ {
		for j := 0; j < values[i].NumField(); j++ {
			if nokeys {
				tag := c.getParam(values[i].Type().Field(j), Fieldtag)
				if tag == nil || tag.iskey {
					continue
				}
			}
			ret[i+1] = append(ret[i+1], values[i].Field(j).String())
		}
	}
	return ret
}

// validateFields is a function that checks if all fields of a struct are strings
func (c *Commands2) validateFields(params []*Param) (map[string]int, error) {
	names := map[string]int{}
	for _, i := range params {
		if i.ftype != "string" {
			return nil, errors.New(ErrorNotStringField)
		}
		for _, name := range i.names {
			names[name]++
		}
	}
	for k, v := range names {
		if v > 1 {
			return nil, fmt.Errorf(ErrorFieldDuplicated, k)
		}
	}
	return names, nil
}

// getTags is a function that returns all tags of a struct
func (c *Commands2) getParams(st reflect.Type, tagname string, args ...string) ([]*Param, error) {
	ret := []*Param{}
	keys := slices.Contains(args, "keys")
	for i := 0; i < st.NumField(); i++ {
		tag := c.getParam(st.Field(i), tagname)
		if tag == nil {
			continue
		}
		if keys && !tag.iskey {
			continue
		}
		ret = append(ret, tag)
	}
	return ret, nil
}

// getTag is a function that returns a tag struct of a struct
func (c *Commands2) getParam(field reflect.StructField, tagname string) *Param {
	tag := field.Tag.Get(tagname)
	if tag == "" {
		return nil
	}
	names, notnull, iskey, pos := c.getParamValues(tag)
	return &Param{names: names, name: "", field: field.Name, ftype: field.Type.String(), 
	              notnull: notnull, iskey: iskey, pos: pos}
}

// splitValues is a function that splits the values of a tag into name, notnull and iskey
func (c *Commands2) getParamValues(tag string) ([]string, bool, bool, string) {
	fields := strings.Split(tag, ";")
	names := []string{}
	notnull := false
	iskey := false
	position := ""
	for _, fd := range fields {
		if strings.Contains(fd, Tagname) {
			names = c.getNames(fd)
		} else if strings.Contains(fd, Tagnotnull) {
			notnull = true
		} else if strings.Contains(fd, Tagkey) {
			iskey = true
		} else if strings.Contains(fd, TagPos) {
			position = c.getPositions(fd)
		}
	}
	return names, notnull, iskey, position
}

// getNames is a function that returns the names inside a tag
func (c *Commands2) getNames(data string) []string {
	ret := []string{}
	s := strings.Split(data, ":")
	if len(s) != 2 || s[0] == Tagname || s[1] == "" {
		return ret
	}
	ret = strings.Split(s[1], ",")
	for i, name := range ret {
		ret[i] = strings.TrimSpace(name)
	}
	return ret
}

// getPositions is a function that returns the positions inside a tag
func (c *Commands2) getPositions(data string) string {
	s := strings.Split(data, ":")
	if len(s) != 2 || s[0] == TagPos || s[1] == "" {
		return ""
	}
	return strings.TrimSpace(s[1])
}

// mapValues is a function that maps values to a struct
func (c *Commands2) mapValues(data string, params []*Param, names map[string]int) error {
	values := strings.Split(data, " ")
	message := ""
	for _, param := range params {
		vals := c.posValues(param.pos, values)
		found := false
		for _, name := range param.names {
			pos := c.index(vals, name)
			if len(pos) > 0 {	
				param.value = c.getValue(pos[0]+1, names, vals)
				param.name = name
				found = true
				if len(pos) > 1 {
					message += fmt.Sprintf(ErrorWordDuplicated, name) + " | "
				}
				break
			}
		}
		if !found && (param.iskey || param.notnull) {
			message += fmt.Sprintf(ErrorKeyNotFound, param.field) + " | "
		} else if param.notnull && param.value == "" {
			message += fmt.Sprintf(ErrorNotNullField, param.field) + " | "
		}
	}
	if message != "" {
		return errors.New(message[:len(message)-3])
	}
	return nil
}

// index is a function that returns the indexes of a string in a slice
func (c *Commands2) index (ss []string, s string) []int {
	ret := []int{}
	i := slices.Index(ss, s)
	for i != -1 {
		ret = append(ret, i)
		ss = ss[i+1:]
		i = slices.Index(ss, s)
	}
	return ret
}

// posValues is a function that returns the values based on the position places on a tag
func (c *Commands2) posValues(posTag string, values []string) []string {
	if posTag == "" {
		return values
	}
	posType := posTag[len(posTag)-1]
	posVal, err := strconv.ParseInt(posTag[:len(posTag)-1], 10, 64)
	if err != nil {
		return values
	}
	if posType == '+' {
		return values[posVal-1:]
	}
	if posType == '-' {
		return values[:posVal]
	}
	if posType == '.' {
		return []string{values[posVal]}
	}
	return values
}

// getValue is a function that returns the value of a tag
func (c *Commands2) getValue(pos int, names map[string]int, ss []string) string {
	value := ""
	for j := pos; j < len(ss); j++ {
		if _, ok := names[ss[j]]; ok {
			break
		}
		if ss[j][0] == '.' {
			ss[j] = ss[j][1:]
		}
		if ss[j] != "" {
			value += ss[j] + " "
		}
	}
	return strings.TrimSpace(value)
}

// setFields is a function that sets the fields of a DTO
func (c *Commands2) setFields(v interface{}, params []*Param) {
	for _, param := range params {
		field := reflect.ValueOf(v).Elem().FieldByName(param.field)
		if param.iskey {
			field.SetString(param.name)
		} else {
			field.SetString(param.value)
		}
	}
}

// trimTable is a function that trims a table
func (c *Commands2) trimTable(table [][]string) [][]string {
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
func (c *Commands2) mountTable(table [][]string) string {
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
