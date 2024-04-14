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
	isfound string
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
		tags, err := c.getParams(st, Fieldtag, "keys")
		if err != nil {
			return nil, err
		}
		if err := c.checkFields(tags); err != nil {
			return nil, err
		}
		c.mapValues(data, tags)
		if err := c.checkValues(tags); err != nil {
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
	if err := c.checkFields(tags); err != nil {
		return err
	}
	c.mapValues(data, tags)
	if err := c.checkValues(tags); err != nil {
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
			fmt.Println(1, obj.Kind())
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



// checkFieldsType is a function that checks if all fields of a struct are strings
func (c *Commands2) checkFields(params []*Param) error {
	names := map[string]int{}
	for _, i := range params {
		if i.ftype != "string" {
			return errors.New(ErrorNotStringField)
		}
		for _, name := range i.names {
			names[name]++
		}
	}
	for k, v := range names {
		if v > 1 {
			return fmt.Errorf(ErrorFieldDuplicated, k)
		}
	}
	return nil
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
	names, notnull, iskey, pos := c.splitValues(tag)
	return &Param{names: names, name: "", field: field.Name, ftype: field.Type.String(), 
	              notnull: notnull, iskey: iskey, pos: pos}
}

// splitValues is a function that splits the values of a tag into name, notnull and iskey
func (c *Commands2) splitValues(tag string) ([]string, bool, bool, string) {
	fields := strings.Split(tag, ";")
	names := []string{}
	notnull := false
	iskey := false
	position := ""
	for _, fd := range fields {
		if !strings.Contains(fd, Tagname) {
			continue
		}
		s := strings.Split(fd, ":")
		if len(s) != 2 || s[0] == Tagname || s[1] == "" {
			continue
		}
		names = strings.Split(s[1], ",")
		for i, name := range names {
			names[i] = strings.TrimSpace(name)
		}
		if strings.Contains(fd, Tagnotnull) {
			notnull = true
		}
		if strings.Contains(fd, Tagkey) {
			iskey = true
		}
		if strings.Contains(fd, TagPos) {
			s := strings.Split(fd, ":")
			if len(s) == 2 && s[0] == TagPos || s[1] != "" {
				position = strings.TrimSpace(s[1])
			}
		}
	}
	return names, notnull, iskey, position
}

// mapValues is a function that maps values to a struct
func (c *Commands2) mapValues(data string, tags []*Param) {
	ss := strings.Split(data, " ")
	for _, tag := range tags {
		vals := c.posValues(tag.pos, ss)
		tag.isfound = "false"
		for _, name := range tag.names {
			pos := slices.Index(vals, name)
			if pos != -1 {
				tag.value = strings.TrimSpace(c.getValue(pos+1, tags, vals))
				tag.isfound = "true"
				break
			}
		}
	}
}

// posValues is a function that returns the values based on the position places on a tag
func (c *Commands2) posValues(posTag string, ss []string) []string {
	if posTag == "" {
		return ss
	}
	posType := posTag[len(posTag)-1]
	posVal, err := strconv.ParseInt(posTag[:len(posTag)-1], 10, 64)
	if err != nil {
		return ss
	}
	if posType == '+' {
		return ss[posVal-1:]
	}
	if posType == '-' {
		return ss[:posVal]
	}
	if posType == '.' {
		return []string{ss[posVal]}
	}
	return ss
}

// getValue is a function that returns the value of a tag
func (c *Commands2) getValue(pos int, tags []*Param, ss []string) string {
	value := ""
	for j := pos; j < len(ss); j++ {
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
func (c *Commands2) checkValues(tags map[string]*Param) error {
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
func (c *Commands2) setFields(v interface{}, tags map[string]*Param) {
	for _, i := range tags {
		field := reflect.ValueOf(v).Elem().FieldByName(i.field)
		field.SetString(i.value)
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
