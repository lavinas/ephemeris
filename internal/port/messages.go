package port

// Messages is a struct that contains all messages returned by the application
const (
	RoleClient             = "client"
	RoleLiable             = "liable"
	RolePayer              = "payer"
	DefaultContact         = "email"
	Location               = "America/Sao_Paulo"
	DateFormat             = "02/01/2006"
	ErrPrefBadRequest      = "bad request"
	ErrPrefCommandNotFound = "command not identified"
	ErrPrefInternal        = "internal error"
	ErrPrefConflict        = "conflict"
	ErrInvalidDateFormat   = "invalid date format. Use dd/mm/yyyy"
	ErrAlreadyExists       = "register already exists with id %s"
	ErrUnfound             = "registers unfound with the informed params"
	ErrParamsNotInformed   = "no params is informed"
	ErrIdUninformed        = "id is not informed"
	ErrEmptyID             = "empty id"
	ErrLongID              = "id should have at most 25"
	ErrInvalidID           = "id should have just one word. Use _ to separate words"
	ErrEmptyName           = "empty name"
	ErrLongName            = "name should have at most 100"
	ErrInvalidName         = "name should have at least two words"
	ErrLongResponsible     = "responsible should have at most 100"
	ErrInvalidResponsible  = "responsible should have at least two words"
	ErrEmptyEmail          = "empty email"
	ErrInvalidEmail        = "invalid email"
	ErrLongEmail           = "email should have at most 100"
	ErrEmptyPhone          = "empty phone"
	ErrLongPhone           = "phone should have at most 20"
	ErrInvalidPhone        = "invalid phone"
	ErrEmptyContact        = "empty contact"
	ErrLongContact         = "contact should have at most 20"
	ErrInvalidContact      = "invalid contact"
	ErrInvalidDocument     = "invalid document"
	ErrLongDocument        = "document should have at most 20"
	ErrInvalidTypeOnMerge  = "invalid type on merge structures"
	ErrCommandNotFound     = "command not identified. Please, see the help command"
	ErrClientIDNotProvided = "client id not provided"
	ErrRoleNotProvided     = "role not provided"
	ErrRefIDNotProvided    = "cliente referece id not provided"
	ErrInvalidRole         = "invalid role. Should be client, liable or payer"
	ErrInvalidReference    = "reference should be different from client"
	ErrLongClientID        = "client id should have at most 25"
	ErrInvalidClientID     = "invalid client id"
	ErrLongRefID           = "ref id should have at most 25"
	ErrInvalidRefID        = "invalid ref id"
	ErrDuplicatedRole      = "this connection between clients already exists"
	ErrSameClient          = "client and reference should be different"
	ErrRefNotFound         = "reference not found or reference is not a client"
	ErrClientNotFound      = "client not found"
	ErrInvalidMinutes      = "minutes should be greater than or equal to zero"
	ErrEmptyCycle          = "empty cycle. Shoud be %s"
	ErrInvalidCycle        = "invalid cycle. Shoud be %s"
	ErrInvalidAmount       = "quantity should be greater than zero or equal zero"
	ErrInvalidLimit        = "limit should be greater than zero or equal zero"
	ErrEmptyAmount         = "if cycle is not once, quantity should be greater than zero"
	ErrZeroAmount          = "if cycle is once, quantity should be zero"
	ErrZeroLimit           = "if cycle is once, limit should be zero"
)
