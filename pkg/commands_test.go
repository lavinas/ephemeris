package pkg

import (
	"testing"
)

var (
	s = struct {
		Nono    string
		Name    string `command:"name:#name; key; not null"`
		Age     string `command:"name:#age; key; not null"`
		Mood    string `command:"name:#mood; key"`
		Other   string `command:"name:#other; not null"`
		Another string `command:"name:#another"`
	}{}

	u = struct {
		Name string `command:"name:#name; key; not null"`
		Test string `command:"name:#test; key; not null"`
	}{}

	v = []interface{}{&s, &u}
)

func TestUnmarshallOk(t *testing.T) {
	commands := NewCommands()
	cmd := "#name alex #age 20 #mood test #other test2 #another xxx"
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
		"#name alex #age 20 #mood test #other test2 #another xxx": "",
		"#name alex #age 20 #mood test #other test2 #another":     "",
		"#name alex #age 20 #mood test #other test2":              "",
		"#name alex #age 20 #mood test #other":                    "tag #other is null",
		"#name alex #age 20 #mood test":                           "tag #other not found",
		"#name alex #age 20":                                      "tag #mood not found | tag #other not found",
	}
	commands := NewCommands()
	for k, v := range testMap {
		err := commands.Unmarshal(k, &s)
		if v == "" && err != nil {
			t.Errorf("Expected nil error, got %s", err.Error())
		}
		if v != "" && err == nil {
			t.Errorf("Expected error %s, got nil", v)
		}
		if v != "" && err != nil && err.Error() != v {
			t.Errorf("Expected error %s, got %s", v, err.Error())
		}
	}
}

func TestUnmarshallNotStringField(t *testing.T) {
	s := struct {
		Name    int    `command:"name:#name; key; not null"`
		Age     string `command:"name:#age; key; not null"`
		Mood    string `command:"name:#mood; key"`
		Other   string `command:"name:#other; not null"`
		Another string `command:"name:#another"`
	}{}
	commands := NewCommands()
	err := commands.Unmarshal("#name alex #age 20 #mood test #other test2 #another xxx", &s)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != ErrorNotStringField {
		t.Errorf("Expected error %s, got %s", ErrorNotStringField, err.Error())
	}
}

func TestUnmarshalOne(t *testing.T) {
	// Ok
	commands := NewCommands()
	cmd := "#name alex #test 20"
	i, err := commands.UnmarshalOne(cmd, v)
	if err != nil {
		t.Errorf("Expected nil error, got %s", err.Error())
	}
	if i == nil {
		t.Errorf("Expected struct, got nil")
	}
	if i != &u {
		t.Errorf("Expected struct, got %v", i)
	}
	if u.Name != "alex" {
		t.Errorf("Expected alex, got %s", u.Name)
	}
	if u.Test != "20" {
		t.Errorf("Expected 20, got %s", u.Test)
	}
	//  Not found
	cmd = "#name alex"
	i, err = commands.UnmarshalOne(cmd, v)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if i != nil {
		t.Errorf("Expected nil, got %v", i)
	}
	// Duplicated
	cmd = "#name alex #age 20 #mood #test #other test2 #another xxx"
	_, err = commands.UnmarshalOne(cmd, v)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestMarshal(t *testing.T) {
	commands := NewCommands()
	cmd := "#name alex #age 20 #mood test #other test2 #another xxx"
	s := struct {
		Name    string `command:"name:#name; key; not null"`
		Age     string `command:"name:#age; key; not null"`
		Mood    string `command:"name:#mood; key"`
		Other   string `command:"name:#other; not null"`
		Another string `command:"name:#another"`
	}{}
	err := commands.Unmarshal(cmd, &s)
	if err != nil {
		t.Errorf("Expected nil error, got %s", err.Error())
	}
	if commands.Marshal(&s) != "Name: alex | Age: 20 | Mood: test | Other: test2 | Another: xxx" {
		t.Errorf("Expected Name: alex | Age: 20 | Mood: test | Other: test2 | Another: xxx, got %s", commands.Marshal(&s))
	}
}
