package dto

type InvoiceListRequest struct {
	Page        int    `json:"page" validate:"required,gt=0"`
	PageSize    int    `json:"page_size" validate:"required,gt=0"`
	Customer    string `json:"customer,omitempty"`
	Vendor      string `json:"vendor,omitempty"`
	Status      string `json:"status,omitempty"`
	InvoiceDate string `json:"invoicing" validate:"required"`
	DueDate     string `json:"due" validate:"required"`
}
