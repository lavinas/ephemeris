package pkg

import (
	"testing"
)

func TestOk(t *testing.T) {
	commands := NewCommands()
	var s = struct {
		Name string `command:"name"`
		Age  string `command:"age"`
	}{}
	if err := commands.Unmarshal("name alex age 20", &s, "command"); err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	if s.Name != "alex" {
		t.Errorf("Name is not alex")
	}
	if s.Age != "20" {
		t.Errorf("Age is not 20")
	}
}

func TestErrorEmptyData(t *testing.T) {
	commands := NewCommands()
	var s = struct {
		Name string `command:"name"`
		Age  string `command:"age"`
	}{}
	if err := commands.Unmarshal("", &s, "command"); err == nil {
		t.Errorf("Error is nil")
	}
}
