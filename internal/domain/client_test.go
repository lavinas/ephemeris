package domain


import (
	"testing"
)

func TestValidate(t *testing.T) {
	// Test valid client
	cli := NewClient("1", "John Doe", "john@doe.com", "011980876112", "email", "04417932824")
	err := cli.Validate()
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}
