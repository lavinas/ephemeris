package dto

type InvoiceCrud struct {
	Object     string `json:"-" command:"name:invoice;key;pos:2-"`
	Action     string `json:"-" command:"name:add,get,up;key;pos:2-"`
	ID         string `json:"id" command:"name:id;pos:3+"`
	Date       string `json:"date" command:"name:date;pos:3+"`
	ClientID   string `json:"client_id" command:"name:client;pos:3+"`
	Kind       string `json:"kind" command:"name:kind;pos:3+"`
	Status     string `json:"status" command:"name:status;pos:3+"`
	SendStatus string `json:"send_status" command:"name:send;pos:3+"`
	PaymentStatus string `json:"payment_status" command:"name:payment;pos:3+"`
}