package domain

// Tax represents the tax information associated with an invoice, including the tax date and related details.
type Tax struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}
