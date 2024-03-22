package dto

// ClientAdd represents the dto for adding a client
type ClientAdd struct {
	ID		    string `json:"id" command:"id"`
	Name        string `json:"name" command:"name"`
	Responsible string `json:"responsible" command:"responsible"`
	Email       string `json:"email" command:"email"`
	Phone       string `json:"phone" command:"phone"`
	Contact     string `json:"contact" command:"contact"`
	Document    string `json:"document" command:"document"`
}

// ClientGet represents the dto for getting a client
type ClientGet struct {
	ID string `json:"id" command:"id"`
	Name string `json:"name" command:"name"`
	Responsible string `json:"responsible" command:"responsible"`
	Email string `json:"email" command:"email"`
	Phone string `json:"phone" command:"phone"`
	Contact string `json:"contact" command:"contact"`
	Document string `json:"document" command:"document"`
}