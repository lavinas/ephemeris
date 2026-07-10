package dto

// PixRequest represents the request structure for generating a Pix payment payload.
type PixRequest struct {
	Key         string
	Description string
	Name        string
	City        string
	Amount      float64
	Txid        string
}
