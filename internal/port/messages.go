package port

// Messages is a struct that contains all messages returned by the application
const (
	ErrAlreadyExists      = "register already exists with id %s"
	ErrUnfound            = "registers unfound with the informed params"
	ErrParamsNotInformed  = "no params is informed"
	ErrIdUninformed       = "id is not informed"
	ErrEmptyName          = "empty name"
	ErrLongName           = "name should have at most 100"
	ErrInvalidName        = "name should have at least two words"
	ErrLongResponsible    = "responsible should have at most 100"
	ErrInvalidResponsible = "responsible should have at least two words"
	ErrEmptyEmail         = "empty email"
	ErrInvalidEmail       = "invalid email"
	ErrLongEmail          = "email should have at most 100"
	ErrEmptyPhone         = "empty phone"
	ErrLongPhone          = "phone should have at most 20"
	ErrInvalidPhone       = "invalid phone"
	ErrEmptyContact       = "empty contact"
	ErrLongContact        = "contact should have at most 20"
	ErrInvalidContact     = "invalid contact"
	ErrInvalidDocument    = "invalid document"
	ErrLongDocument       = "document should have at most 20"
	ErrInvalidTypeOnMerge = "invalid type on merge structures"
)
