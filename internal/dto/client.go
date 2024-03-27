package dto

// ClientAdd represents the dto for adding a client
type ClientAdd struct {
	Object      string `json:"-" command:"name:client;key"`
	Action      string `json:"-" command:"name:add;key"`
	ID          string `json:"id" command:"name:id"`
	Name        string `json:"name" command:"name:name"`
	Responsible string `json:"responsible" command:"name:responsible"`
	Email       string `json:"email" command:"name:email"`
	Phone       string `json:"phone" command:"name:phone"`
	Contact     string `json:"contact" command:"name:contact"`
	Document    string `json:"document" command:"name:document"`
}


// ClientGet represents the dto for getting a client
type ClientGet struct {
	Object      string `json:"-" command:"name:client;key"`
	Action      string `json:"-" command:"name:get;key"`
	ID          string `json:"id" command:"name:id"`
	Name        string `json:"name" command:"name:name"`
	Responsible string `json:"responsible" command:"name:responsible"`
	Email       string `json:"email" command:"name:email"`
	Phone       string `json:"phone" command:"name:phone"`
	Contact     string `json:"contact" command:"name:contact"`
	Document    string `json:"document" command:"name:document"`
}
