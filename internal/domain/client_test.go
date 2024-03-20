package domain

import (
	"testing"
)

func TestValidate(t *testing.T) {
	longstring := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."
	TestMap := map[*Client]string{
		NewClient("1", "John Doe", "  ", "john@doe.com", "011980876112", "email", "04417932824"):              "",
		NewClient("1", "  John Doe  ", "  ", " john@doe.com", "  011980876112", "  email  ", "  04417932824"): "",
		NewClient("1", "John Doe", "", "john@doe.com", "011980876112", "email", ""):                           "",
		NewClient("1", "John Doe", "Responsible John", "john@doe.com", "011980876112", "email", ""):           "",
		NewClient("1", "  ", "", "john@doe.com", "011980876112", "email", "04417932824"):                      ErrEmptyName,
		NewClient("1", "John", "", "john@doe.com", "011980876112", "email", "04417932824"):                    ErrInvalidName,
		NewClient("1", longstring, "", "john@doe.com", "011980876112", "email", "04417932824"):                ErrLongName,
		NewClient("1", "John Doe", "John", "john@doe.com", "011980876112", "email", "04417932824"):            ErrInvalidResponsible,
		NewClient("1", "John Doe", longstring, "john@doe.com", "011980876112", "email", "04417932824"):        ErrLongResponsible,
		NewClient("1", "  ", "", "john@doe.com", "011980876112", "email", "04417932824"):                      ErrEmptyName,
		NewClient("1", "John", "", "john@doe.com", "011980876112", "email", "04417932824"):                    ErrInvalidName,
		NewClient("1", longstring, "", "john@doe.com", "011980876112", "email", "04417932824"):                ErrLongName,
		NewClient("1", "John Doe", "", "", "011980876112", "email", "04417932824"):                            ErrEmptyEmail,
		NewClient("1", "John Doe", "", "john", "011980876112", "email", "04417932824"):                        ErrInvalidEmail,
		NewClient("1", "John Doe", "", longstring, "011980876112", "email", "04417932824"):                    ErrLongEmail,
		NewClient("1", "John Doe", "", "john@doe.com", "", "email", "04417932824"):                            ErrEmptyPhone,
		NewClient("1", "John Doe", "", "john@doe.com", "211980876112", "email", "04417932824"):                ErrInvalidPhone,
		NewClient("1", "John Doe", "", "john@doe.com", longstring, "email", "04417932824"):                    ErrLongPhone,
		NewClient("1", "John Doe", "", "john@doe.com", "011980876112", "", "04417932824"):                     ErrEmptyContact,
		NewClient("1", "John Doe", "", "john@doe.com", "011980876112", "phone", "04417932824"):                ErrInvalidContact,
		NewClient("1", "John Doe", "", "john@doe.com", "011980876112", longstring, "04417932824"):             ErrLongContact,
		NewClient("1", "John Doe", "", "john@doe.com", "011980876112", "email", "84417932824"):                ErrInvalidDocument,
		NewClient("1", "John Doe", "", "john@doe.com", "011980876112", "email", longstring):                   ErrLongDocument,
	}
	for k, v := range TestMap {
		err := k.Validate()
		if err == nil && v == "" {
			continue
		}
		if err != nil && err.Error() == v {
			continue
		}
		t.Errorf("Expected %v, got %v", v, err)
	}
}

func TestFormat(t *testing.T) {
	c := NewClient("1", " John dOe silva ", " resPonsible John silva", "  john@doe.com  ", "011980876112", " email ", "04417932824")
	c.Format()
	if c.Name != "John Doe Silva" {
		t.Errorf("Expected John Doe Silva, got %v", c.Name)
	}
	if c.Responsible != "Responsible John Silva" {
		t.Errorf("Expected Responsible John Silva, got %v", c.Responsible)
	}
	if c.Email != "john@doe.com" {
		t.Errorf("Expected john@doe.com, got %v", c.Email)
	}
	if c.Phone != "+5511980876112" {
		t.Errorf("Expected +5511980876112, got %v", c.Phone)
	}
	if c.Contact != "email" {
		t.Errorf("Expected email, got %v", c.Contact)
	}
	if c.Document != "044.179.328-24" {
		t.Errorf("Expected 044.179.328-24, got %v", c.Document)
	}
}
