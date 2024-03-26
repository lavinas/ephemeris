package dto

// ClientAdd represents the dto for adding a client
type ClientAdd struct {
	Object      string `json:"-" command:"client; key"`
	Action      string `json:"-" command:"add; key"`
	ID		    string `json:"id" command:"id; not null"`
	Name        string `json:"name" command:"name; nor null"`
	Responsible string `json:"responsible" command:"responsible"`
	Email       string `json:"email" command:"email; not null"`
	Phone       string `json:"phone" command:"phone; not null"`
	Contact     string `json:"contact" command:"contact"`
	Document    string `json:"document" command:"document"`
}

// ClientGet represents the dto for getting a client
type ClientGet struct {
	Base
	Name        string `json:"name" command:"name"`
	Responsible string `json:"responsible" command:"responsible"`
	Email       string `json:"email" command:"email"`
	Phone       string `json:"phone" command:"phone"`
	Contact     string `json:"contact" command:"contact"`
	Document    string `json:"document" command:"document"`
}
