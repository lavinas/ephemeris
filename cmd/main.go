package main

import (
	"fmt"
	"reflect"
	"slices"
)

const (
	ErrorMissingParameter = "missing parameter: %s"
)

type Base struct {
	Object string `json:"object" command:"object"`
	Action string `json:"action" command:"action"`
	ID     string `json:"id" command:"id"`
}

func (b *Base) GetObject() string {
	return b.Object
}

func (b *Base) GetAction() string {
	return b.Action
}

type DTO interface {
	GetObject() string
	GetAction() string
}

type ClientAdd struct {
	Base
	Name        string `json:"name" command:".name"`
	Responsible string `json:"responsible" command:".responsible"`
}

func main() {
	clientAdd := ClientAdd{
		Base: Base{
			Object: "client",
			Action: "add",
			ID:     "1",
		},
		Name:        "client name",
		Responsible: "responsible name",
	}
	command2Params(&clientAdd, []string{".client", ".add", ".name", "Paulo", "Barbosa", ".responsible", "Joao", "da", "Silva"})
	fmt.Println(clientAdd)
}

// Command2DTO is a function that converts a string command to a DTO
func command2Params(dto DTO, cmdSlice []string) error {
	st := reflect.TypeOf(dto).Elem()
	tags := commandGetTags(st)
	values, err := commandMapValues(tags, cmdSlice)
	if err != nil {
		return err
	}
	commandSetFields(dto, values)
	return nil
}

// getCommandTags is a function that returns all tags of a struct
func commandGetTags(st reflect.Type) map[string]string {
	tags := map[string]string{}
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag.Get("command")
		if tag == "" {
			continue
		}
		tags[tag] = st.Field(i).Name
	}
	return tags
}

// mapCommandValues is a function that maps command fields values
func commandMapValues(alltags map[string]string, cmdSlice []string) (map[string]string, error) {
	valueMap := map[string]string{}
	for tag, field := range alltags {
		if !slices.Contains(cmdSlice, tag) {
			return valueMap, fmt.Errorf(ErrorMissingParameter, tag)
		}
		param := ""
		for j := slices.Index(cmdSlice, tag) + 1; j < len(cmdSlice); j++ {
			if alltags[cmdSlice[j]] != "" {
				break
			}
			param += cmdSlice[j] + " "
		}
		valueMap[field] = param
	}
	return valueMap, nil
}

// setCommandFields is a function that sets the fields of a DTO
func commandSetFields(dto DTO, values map[string]string) {
	for k, v := range values {
		field := reflect.ValueOf(dto).Elem().FieldByName(k)
		field.SetString(v)
	}

}
