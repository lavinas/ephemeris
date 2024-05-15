package dto

// AgendaReport represents the dto for reporting a agenda
type AgendaReport struct {
	Object     string `json:"-" command:"name:agenda;key;pos:2-"`
	Action     string `json:"-" command:"name:report,repo,rep;key;pos:2-"`
	ClientID   string `json:"client" command:"name:client;pos:3+"`
	ContractID string `json:"contract" command:"name:contract;pos:3+"`
	At         string `json:"at" command:"name:at;pos:3+"`
	Minutes    string `json:"minutes" command:"name:minutes,min;pos:3+"`
	Kind       string `json:"kind" command:"name:kind;pos:3+"`
}
