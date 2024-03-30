package pkg

// Messages is a struct that contains all messages returned by the application
const (
	ErrorStructDuplicated  = "command words are duplicated in the struct"
	ErrorCommandDuplicated = "more than one command found. Try use . in front of the parameter words if parameter has command words"
	ErrorCommandNotFound   = "command not found with the given parameters"
	ErrorWordDuplicated    = "command word(s) %s are duplicated. Try use . in front of the parameter words if parameter has command words"
	ErrorTagNameNotFound   = "tag name not found"
	ErrorNotStringField    = "not all fields are strings"
	ErrorKeyNotFound       = "tag %s not found"
	ErrorNotNullField      = "tag %s is null"
	Fieldtag               = "command"
	Tagname                = "name:"
	Tagnotnull             = "not null"
	Tagkey                 = "key"
)
