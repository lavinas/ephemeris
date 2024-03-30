package pkg

import (
	"testing"
)

var (
	s = struct {
		Nono    string
		Non2    string `command:"key; not null"`
		Name    string `command:"name:name; key; not null"`
		Age     string `command:"name:age; key; not null"`
		Mood    string `command:"name:mood; key"`
		Other   string `command:"name:other; not null"`
		Another string `command:"name:another"`
	}{}

	x = struct {
		Name  string `command:"name:name; key; not null"`
		Name2 string `command:"name:name; key; not null"`
	}{}
)

func TestUnmarshallOk(t *testing.T) {
	commands := NewCommands()
	cmd := "name alex age 20 mood test other test2 another xxx"
	err := commands.Unmarshal(cmd, &s)
	if err != nil {
		t.Errorf("Expected nil error, got %s", err.Error())
	}
	if s.Name != "alex" {
		t.Errorf("Expected alex, got %s", s.Name)
	}
	if s.Age != "20" {
		t.Errorf("Expected 20, got %s", s.Age)
	}
	if s.Mood != "test" {
		t.Errorf("Expected test, got %s", s.Mood)
	}
	if s.Other != "test2" {
		t.Errorf("Expected test2, got %s", s.Other)
	}
	if s.Another != "xxx" {
		t.Errorf("Expected xxx, got %s", s.Another)
	}
}

func TestUnmarshallComplete(t *testing.T) {
	testMap := map[string]string{
		"name alex age 20 mood . .test other test2 another xxx": "",
		"name alex age 20 mood test other test2 another":        "",
		"name alex age 20 mood test other test2":                "",
		"name alex age 20 mood test other":                      "tag other is null",
		"name alex age 20 mood test":                            "tag other not found",
		"name alex age 20":                                      "tag mood not found | tag other not found",
		"name name age 20 mood test other test2 another xxx":    "command word(s) name are duplicated. Try use . in front of the parameter words if parameter has command words",
	}
	commands := NewCommands()
	for k, v := range testMap {
		err := commands.Unmarshal(k, &s)
		if v == "" && err != nil {
			t.Errorf("Expected nil error, got: %s", err.Error())
		}
		if v != "" && err == nil {
			t.Errorf("Expected error: %s, got: nil", v)
		}
		if v != "" && err != nil && err.Error() != v {
			t.Errorf("Expected error: %s, got: %s", v, err.Error())
		}
	}
}

func TestUnmarshallStructError(t *testing.T) {
	commands := NewCommands()
	err := commands.Unmarshal("name alex age 20 mood test other test2 another xxx", &x)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err != nil && err.Error() != ErrorStructDuplicated {
		t.Errorf("Expected error: %s, got: %s", ErrorStructDuplicated, err.Error())
	}
}

func TestUnmarshallNotStringField(t *testing.T) {
	s := struct {
		Name    int    `command:"name:name; key; not null"`
		Age     string `command:"name:age; key; not null"`
		Mood    string `command:"name:mood; key"`
		Other   string `command:"name:other; not null"`
		Another string `command:"name:another"`
	}{}
	commands := NewCommands()
	err := commands.Unmarshal("name alex age 20 mood test other test2 another xxx", &s)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != ErrorNotStringField {
		t.Errorf("Expected error: %s, got: %s", ErrorNotStringField, err.Error())
	}
}

func TestMarshal(t *testing.T) {
	commands := NewCommands()
	cmd := "name alex age 20 mood test other test2 another xxx"
	s := struct {
		Name    string `command:"name:name; key; not null"`
		Age     string `command:"name:age; key; not null"`
		Mood    string `command:"name:mood; key"`
		Other   string `command:"name:other; not null"`
		Another string `command:"name:another"`
	}{}
	err := commands.Unmarshal(cmd, &s)
	if err != nil {
		t.Errorf("Expected: nil error, got: %s", err.Error())
	}
	if commands.Marshal(&s) == "" {
		t.Errorf("\nExpected:\n%s\ngot:\n%s", "", commands.Marshal(&s))
	}
}
