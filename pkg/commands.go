package pkg

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Command is a struct that represents a command
type Param struct {
	names     []string
	name      string
	field     string
	ftype     string
	iskey     bool
	notnull   bool
	posInit   *int
	posEnd    *int
	corr      []string
	value     string
	transpose string
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
		return ErrNoResults
	}
	nokeys := slices.Contains(args, "nokeys")
	counter := slices.Contains(args, "counter")
	ret := c.getValuesSlice(rvl, nokeys, counter)
	if len(ret) == 0 {
		return ErrNoResults
	}
	if slices.Contains(args, "trim") {
		ret = c.trimTable(ret)
	}
	if slices.Contains(args, "more") {
		cols := len(ret[0])
		more := []string{}
		for i := 0; i < cols; i++ {
			more = append(more, "...")
		}
		ret = append(ret, more)
	}
	return c.mountTable(ret)
}

// UnmarshalOne is a function that converts a string to a struct
func (c *Commands) FindOne(data string, v []interface{}) (interface{}, error) {
	ret := []interface{}{}
	for _, i := range v {
		st := reflect.TypeOf(i).Elem()
		params, err := c.getParams(st, Fieldtag, "keys")
		if err != nil {
			return nil, err
		}
		if err := c.validateFields(params); err != nil {
			return nil, err
		}
		if err := c.mapValues(data, params); err != nil {
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
func (c *Commands) Unmarshal(data string, v interface{}) error {
	st := reflect.TypeOf(v).Elem()
	params, err := c.getParams(st, Fieldtag)

	if err != nil {
		return err
	}
	if err := c.validateFields(params); err != nil {
		return err
	}
	if err := c.mapValues(data, params); err != nil {
		return err
	}
	c.setFields(v, params)
	return nil
}

// Transpose is a function that returns the transpose of a struct in a slice of strings
// transpose string, numeric and time fields with tag command and sub-tab transpose
func (c *Commands) Transpose(v interface{}) ([]interface{}, error) {
	trans := []interface{}{}
	etype := reflect.TypeOf(v).Elem()
	eval := reflect.ValueOf(v).Elem()
	params := []*Param{}
	for i := 0; i < etype.NumField(); i++ {
		param := c.getParam(etype.Field(i), Fieldtag)
		if param == nil {
			continue
		}
		val, trs, err := c.transpose(eval.Field(i).String(), param)
		if err != nil {
			return nil, err
		}
		param.value = val
		if param.iskey {
			param.name = param.value
		}
		params = append(params, param)
		if trs != "" {
			trans = append(trans, trs)
		}
	}
	c.setFields(v, params)
	return trans, nil
}

// Order is a function that orders a slice of structs by a field
func (c *Commands) Sort(v interface{}, cmd string) {
	if cmd == "" {
		return
	}
	field, down := c.sortParams(cmd)
	if field == "" {
		return
	}
	c.sort(v, field, down)
}

// WeightedEuclidean returns the euclidean distance between x and y with weights w
// and w is a slice of weights for each character in x and y
// if w is nil, the weights are calculated as a decreasing sequence
// starting in 1 and ending in 1/len(x)
// returns the euclidean distance between x and y with weights w or an error
func (c *Commands) WeightedEuclidean(x, y string, w []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, fmt.Errorf("x, y must have the same length")
	}
	if w != nil && len(w) < len(x) {
		return 0, fmt.Errorf("w must be at least the same length as x")
	}
	k := 1.00000000
	dec := 1 / float64(len(x))
	z := 0.00000000
	for i := 0; i < len(x); i++ {
		z1 := float64(x[i] - y[i])
		z1 = z1 * z1
		if w != nil {
			z1 = z1 * w[i]
		} else {
			z1 = z1 * k
			k -= dec
		}
		z += z1
	}
	return math.Sqrt(z), nil
}

// sortParams is a function that returns the field and the direction of a command
func (c *Commands) sortParams(cmd string) (string, bool) {
	field := strings.Split(cmd, " ")[0]
	down := false
	if strings.Contains(cmd, "down") {
		down = true
	}
	if len(field) != 0 && field[0] == '.' {
		field = field[1:]
	}
	return field, down
}

// sort is a function that sorts a slice of structs by a field
func (c *Commands) sort(v interface{}, field string, down bool) {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		v = reflect.ValueOf(v).Elem().Interface()
	}
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		return
	}
	sort.Slice(v, func(i, j int) bool {
		iv := reflect.ValueOf(v).Index(i).Elem().Interface()
		jv := reflect.ValueOf(v).Index(j).Elem().Interface()
		ivf := c.fieldByTag(iv, field)
		jvf := c.fieldByTag(jv, field)
		if down {
			return ivf > jvf
		}
		return ivf < jvf
	})
}

// fieldByTag is a function that returns the field of a struct by a tag
func (c *Commands) fieldByTag(v interface{}, field string) string {
	etype := reflect.TypeOf(v).Elem()
	eval := reflect.ValueOf(v).Elem()
	for i := 0; i < etype.NumField(); i++ {
		param := c.getParam(etype.Field(i), Fieldtag)
		if param == nil {
			continue
		}
		for _, j := range param.names {
			if j == field {
				return eval.Field(i).String()
			}
		}
	}
	return ""
}

// transpose is a function that returns the transpose of a string
func (c *Commands) transpose(data string, param *Param) (string, string, error) {
	if data == "" || param.transpose == "" {
		return data, "", nil
	}
	trs := strings.Split(param.transpose, ",")
	if len(trs) != 2 {
		return "", "", errors.New(ErrorTransposeStruct)
	}
	field := trs[0]
	ftype := trs[1]
	switch ftype {
	case "string":
		return c.transposeString(data, field)
	case "numeric":
		return c.transposeNumeric(data, field)
	case "time":
		return c.transposeTime(data, field)
	default:
		return "", "", errors.New(ErrorTransposeType)
	}
}

// transposeString is a function that returns the transpose of a string
func (c *Commands) transposeString(data string, field string) (string, string, error) {
	if data == "cmd" {
		data = "c*"
	}
	if !strings.Contains(data, "+") && !strings.Contains(data, "-") && !strings.Contains(data, "*") {
		return data, "", nil
	}
	data = strings.ReplaceAll(data, "*", "%")
	return "", fmt.Sprintf("%s like '%s'", field, data), nil
}

// transposeFloat is a function that returns the transpose of a float
func (c *Commands) transposeNumeric(data string, field string) (string, string, error) {
	data = strings.ReplaceAll(data, ",", ".")
	cmd := data[len(data)-1:]
	switch cmd {
	case "+":
		data = data[:len(data)-1]
		return "", fmt.Sprintf("%s >= %s", field, data), nil
	case "-":
		data = data[:len(data)-1]
		return "", fmt.Sprintf("%s <= %s", field, data), nil
	case "*":
		data = strings.ReplaceAll(data, "*", "%")
		return "", fmt.Sprintf("%s like '%s'", field, data), nil
	default:
		return data, "", nil
	}
}

// transposeTime is a function that returns the transpose of a time
func (c *Commands) transposeTime(data string, field string) (string, string, error) {
	d, ok := c.translateTime(data[:len(data)-1])
	if !ok {
		return data, "", nil
	}
	switch data[len(data)-1:] {
	case "+":
		return "", fmt.Sprintf("%s >= '%s'", field, d), nil
	case "-":
		return "", fmt.Sprintf("%s <= '%s'", field, d), nil
	case "d":
		start := d[:10] + " 00:00:00"
		end := d[:10] + " 23:59:59"
		return "", fmt.Sprintf("%s >= '%s'and %s <= '%s'", field, start, field, end), nil
	case "m":
		start := d[:8] + "01 00:00:00"
		e, _ := time.Parse("2006-01-02 15:04:05", start)
		e = e.AddDate(0, 1, 0)
		d = e.Format("2006-01-02 15:04:05")
		return "", fmt.Sprintf("%s >= '%s'and %s < '%s'", field, start, field, d), nil
	default:
		return data, "", nil
	}
}

// translateTime is a function that translates a time string layout to a default time layout
func (c *Commands) translateTime(data string) (string, bool) {
	fSlice := []string{
		DateTimeFormat,
		DateHourFormat,
		DateFormat,
		MonthFormat,
	}
	for _, v := range fSlice {
		t, err := time.Parse(v, data)
		if err == nil {
			return t.Format("2006-01-02 15:04:05"), true
		}
	}
	return "", false
}


// getInputSlice is a function that returns a slice of reflect.Values
func (c *Commands) getInputSlice(v interface{}) []reflect.Value {
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
func (c *Commands) getValuesSlice(values []reflect.Value, nokeys bool, counter bool) [][]string {
	ret := make([][]string, len(values)+1)
	filled := false
	count := 1
	if counter {
		ret[0] = append(ret[0], "#")
	}
	for i := 0; i < values[0].NumField(); i++ {
		if nokeys {
			tag := c.getParam(values[0].Type().Field(i), Fieldtag)
			if tag == nil || tag.iskey {
				continue
			}
		}
		filled = true
		p, _, _, _, _, _ := c.getParamValues(values[0].Type().Field(i).Tag.Get(Fieldtag))
		ret[0] = append(ret[0], strings.Join(p, ", "))
	}
	for i := 0; i < len(values); i++ {
		if counter {
			ret[i+1] = append(ret[i+1], strconv.Itoa(count))
			count++
		}
		for j := 0; j < values[i].NumField(); j++ {
			if nokeys {
				tag := c.getParam(values[i].Type().Field(j), Fieldtag)
				if tag == nil || tag.iskey {
					continue
				}
			}
			filled = true
			ret[i+1] = append(ret[i+1], values[i].Field(j).String())
		}
	}
	if !filled {
		return [][]string{}
	}
	return ret
}

// validateFields is a function that checks if all fields of a struct are strings
func (c *Commands) validateFields(params []*Param) error {
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
func (c *Commands) getParams(st reflect.Type, tagname string, args ...string) ([]*Param, error) {
	ret := []*Param{}
	keys := slices.Contains(args, "keys")
	for i := 0; i < st.NumField(); i++ {
		params := c.getParam(st.Field(i), tagname)
		if params == nil {
			continue
		}
		if keys && !params.iskey {
			continue
		}
		ret = append(ret, params)
	}
	c.setCorrelations(ret)
	return ret, nil
}

// getTag is a function that returns a tag struct of a struct
func (c *Commands) getParam(field reflect.StructField, tagname string) *Param {
	tag := field.Tag.Get(tagname)
	if tag == "" {
		return nil
	}
	names, notnull, iskey, init, end, transp := c.getParamValues(tag)
	return &Param{names: names, name: "", field: field.Name, ftype: field.Type.String(),
		notnull: notnull, iskey: iskey, posInit: init, posEnd: end, transpose: transp}
}

// splitValues is a function that splits the values of a tag into name, notnull and iskey
func (c *Commands) getParamValues(tag string) ([]string, bool, bool, *int, *int, string) {
	fields := strings.Split(tag, ";")
	names := []string{}
	notnull := false
	iskey := false
	var init *int
	var end *int
	transp := ""
	for _, fd := range fields {
		vals := strings.Split(fd, ":")
		switch vals[0] {
		case Tagname:
			names = c.getNames(vals)
		case Tagnotnull:
			notnull = true
		case Tagkey:
			iskey = true
		case TagPos:
			init, end = c.getPositions(vals)
		case TagTranspose:
			transp = c.getTranspose(vals)
		}
	}
	return names, notnull, iskey, init, end, transp
}

// getNames is a function that returns the names inside a tag
func (c *Commands) getNames(s []string) []string {
	ret := []string{}
	if len(s) != 2 || s[0] != Tagname || s[1] == "" {
		return ret
	}
	ret = strings.Split(s[1], ",")
	for i, name := range ret {
		ret[i] = strings.TrimSpace(name)
	}
	return ret
}

// getPositions is a function that returns the positions inside a tag
func (c *Commands) getPositions(s []string) (*int, *int) {
	if len(s) != 2 || s[0] != TagPos || s[1] == "" {
		return nil, nil
	}
	pos := strings.TrimSpace(s[1])
	signal := pos[len(pos)-1]
	val, err := strconv.Atoi(pos[:len(pos)-1])
	if err != nil {
		return nil, nil
	}
	switch signal {
	case '+':
		return &val, nil
	case '-':
		return nil, &val
	case '.':
		return &val, &val
	default:
		return nil, nil
	}
}

// getTranspose is a function that returns the transpose inside a tag
func (c *Commands) getTranspose(s []string) string {
	if len(s) != 2 || s[0] != TagTranspose || s[1] == "" {
		return ""
	}
	return s[1]
}

// setCorrelations is a function that sets the correlations of a struct
func (c *Commands) setCorrelations(params []*Param) {
	for i := 0; i < len(params); i++ {
		if params[i].posInit == nil && params[i].posEnd == nil {
			continue
		}
		init1 := 0
		if params[i].posInit != nil {
			init1 = *params[i].posInit
		}
		end1 := math.MaxInt
		if params[i].posEnd != nil {
			end1 = *params[i].posEnd
		}
		params[i].corr = append(params[i].corr, params[i].names...)
		for j := i + 1; j < len(params); j++ {
			if params[j].posInit == nil && params[j].posEnd == nil {
				continue
			}
			init2 := 0
			if params[j].posInit != nil {
				init2 = *params[j].posInit
			}
			end2 := math.MaxInt
			if params[j].posEnd != nil {
				end2 = *params[j].posEnd
			}
			if (init1 >= init2 && init1 <= end2) ||
				(end1 >= init2 && end1 <= end2) ||
				(init2 >= init1 && init2 <= end1) ||
				(end2 >= init1 && end2 <= end1) {
				params[i].corr = append(params[i].corr, params[j].names...)
				params[j].corr = append(params[j].corr, params[i].names...)

			}
		}
	}
}

// mapValues is a function that maps values to a struct
func (c *Commands) mapValues(data string, params []*Param) error {
	values := strings.Split(data, " ")
	message := ""
	for _, param := range params {
		vals := c.posValues(param.posInit, param.posEnd, values)
		found := false
		for _, name := range param.names {
			pos := c.index(vals, name)
			if len(pos) > 0 {
				param.value = c.getValue(pos[0]+1, param.corr, vals)
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
func (c *Commands) index(ss []string, s string) []int {
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
func (c *Commands) posValues(init *int, end *int, values []string) []string {
	if init != nil {
		values = values[*init-1:]
	}
	if end != nil {
		if *end > len(values) {
			*end = len(values)
		}
		values = values[:*end]
	}
	return values
}

// getValue is a function that returns the value of a tag
func (c *Commands) getValue(pos int, names []string, ss []string) string {
	value := ""
	for j := pos; j < len(ss); j++ {
		if slices.Contains(names, ss[j]) {
			break
		}
		v := ss[j]
		if v[0] == '.' {
			v = v[1:]
		}
		if v != "" {
			value += ss[j] + " "
		}
	}
	if value == "" {
		return value
	}
	return strings.TrimSpace(value[:len(value)-1])
}

// setFields is a function that sets the fields of a DTO
func (c *Commands) setFields(v interface{}, params []*Param) {
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
