package dto

// BillerItem represents an item in the biller request.
type BillerItem struct {
	Description string
	Quantity    int
	Price       float64
}

// BillerVendor represents the vendor information in the biller request.
type BillerVendor struct {
	Logo     string
	Name     string
	Document string
	Address  *string
	Postcode *string
	City     *string
	State    *string
	Country  *string
	Email    *string
	Whatsapp *string
	PixKey   *string
}

// BillerCustomer represents the customer information in the biller request.
type BillerCustomer struct {
	Name     string
	Document *string
	Email    *string
	Whatsapp *string
}

// BillerRequest defines the interface for generating PDF files.
type BillerRequest struct {
	Vendor   BillerVendor
	Customer BillerCustomer
	Items    []BillerItem
	Notes    *string
}
