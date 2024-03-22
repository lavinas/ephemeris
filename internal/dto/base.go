package dto

// Base represents the base for DTO
type Base struct {
	Object string `json:"object"`
	Action string `json:"action"`
	ID     string `json:"id" command:"id"`
}

// NewBase is a function that creates a new Base for DTO
func NewBase(object string, action string, id string) *Base {
	return &Base{
		Object: object,
		Action: action,
		ID:     id,
	}
}

// GetObject is a method that returns the object of the Base
func (b *Base) GetObject() string {
	return b.Object
}

// GetAction is a method that returns the action of the Base
func (b *Base) GetAction() string {
	return b.Action
}


