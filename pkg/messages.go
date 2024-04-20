package pkg

// Messages is a struct that contains all messages returned by the application
const (
	ErrorStructDuplicated      = "command words are duplicated in the struct"
	ErrorCommandDuplicated     = "more than one command found. Try use . in front of the parameter words if parameter has command words"
	ErrorCommandNotFound       = "command not found with the given parameters"
	ErrorWordDuplicated        = "command word(s) %s are duplicated. Try use . in front of the parameter words if parameter has command words"
	ErrorTagNameNotFound       = "tag name not found"
	ErrorNotStringField        = "not all fields are strings"
	ErrorFieldDuplicated       = "field %s is duplicated in the struct"
	ErrorKeyNotFound           = "tag %s not found"
	ErrorNotNullField          = "tag %s is null"
	Fieldtag                   = "command"
	Tagname                    = "name"
	Tagnotnull                 = "not null"
	Tagkey                     = "key"
	TagPos                     = "pos"
	RoleClient                 = "client"
	RoleLiable                 = "liable"
	RolePayer                  = "payer"
	DefaultContact             = "e-mail"
	DefaultDueDay              = "10"
	DefaultBillingType         = "pre-paid"
	Location                   = "America/Sao_Paulo"
	DateFormat                 = "02/01/2006"
	MonthFormat                = "01/2006"
	ErrPrefBadRequest          = "bad request"
	ErrPrefCommandNotFound     = "command not identified"
	ErrPrefInternal            = "internal error"
	ErrPrefConflict            = "conflict"
	ErrInvalidDateFormat       = "invalid date format. Use dd/mm/yyyy"
	ErrAlreadyExists           = "register already exists with id %s"
	ErrUnfound                 = "registers unfound with the informed params"
	ErrParamsNotInformed       = "no params is informed"
	ErrIdUninformed            = "id is not informed"
	ErrEmptyID                 = "empty id"
	ErrLongID                  = "id should have at most 25"
	ErrInvalidID               = "id should have just one word. Use _ to separate words"
	ErrEmptyName               = "empty name"
	ErrLongName                = "name should have at most 100"
	ErrInvalidName             = "name should have at least two words"
	ErrLongResponsible         = "responsible should have at most 100"
	ErrInvalidResponsible      = "responsible should have at least two words"
	ErrEmptyEmail              = "empty email"
	ErrInvalidEmail            = "invalid email"
	ErrLongEmail               = "email should have at most 100"
	ErrEmptyPhone              = "empty phone"
	ErrLongPhone               = "phone should have at most 20"
	ErrInvalidPhone            = "invalid phone"
	ErrEmptyContact            = "empty contact"
	ErrLongContact             = "contact should have at most 20"
	ErrInvalidContact          = "invalid contact. Should be %s"
	ErrInvalidDocument         = "invalid document"
	ErrLongDocument            = "document should have at most 20"
	ErrInvalidTypeOnMerge      = "invalid type on merge structures"
	ErrCommandNotFound         = "command not identified. Please, see the help command"
	ErrClientIDNotProvided     = "client id not provided"
	ErrRoleNotProvided         = "role not provided"
	ErrRefIDNotProvided        = "cliente referece id not provided"
	ErrInvalidRole             = "invalid role. Should be client, liable or payer"
	ErrInvalidReference        = "reference should be different from client"
	ErrLongClientID            = "client id should have at most 25"
	ErrInvalidClientID         = "invalid client id"
	ErrLongRefID               = "ref id should have at most 25"
	ErrInvalidRefID            = "invalid ref id"
	ErrDuplicatedRole          = "this connection between clients already exists"
	ErrSameClient              = "client and reference should be different"
	ErrRefNotFound             = "reference not found or reference is not a client"
	ErrClientNotFound          = "client not found"
	ErrInvalidMinutes          = "minutes should be greater than or equal to zero"
	ErrEmptyCycle              = "empty cycle. Shoud be %s"
	ErrInvalidCycle            = "invalid cycle. Shoud be %s"
	ErrInvalidAmount           = "quantity should be greater than zero or equal zero"
	ErrInvalidLimit            = "limit should be greater than zero or equal zero"
	ErrEmptyLen                = "if cycle is not once, len should be numeric and greater than zero"
	ErrZeroLen                 = "if cycle is once, len should be zero or not be informed"
	ErrZeroLimit               = "if cycle is once, limit should be zero or not be informed"
	ErrLongCycle               = "cycle should have at most 20"
	ErrEmptyUnitAndPack        = "unit or pack should be informed"
	ErrDuplicityUnitAndPack    = "unit and pack should not be informed at the same time"
	ErrServiceIDNotProvided    = "service id not provided"
	ErrServiceNotFound         = "service not found"
	ErrRecurrenceIDNotProvided = "recurrence id not provided"
	ErrRecurrenceNotFound      = "recurrence not found"
	ErrPriceIDNotProvided      = "price id not provided"
	ErrPriceNotFound           = "price not found"
	ErrEmptyBillingType        = "empty billing type"
	ErrInvalidBillingType      = "invalid billing type. Should be %s"
	ErrInvalidDueDay           = "due day should be until 31 or zero"
	ErrInvalidStartDate        = "invalid start date. Should have %s format"
	ErrInvalidEndDate          = "invalid end date. Should have %s format"
	ErrBondNotFound            = "bond contract not found"
	ErrDueDayNotProvided       = "due day not provided"
	ErrSponsorNotFound         = "sponsor not found"
	ErrPackageNotFound         = "package not found"
	ErrEmptyContractID         = "empty contract id"
	ErrContractNotFound        = "contract not found"
	ErrInvalidEnd              = "end date should be after start date"
)
