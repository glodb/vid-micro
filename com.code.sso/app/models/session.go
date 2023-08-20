package models

//This structure mainly handle the session
/*
registrationType variable can have following values
1- System Registration
2- Google Registration
*/
type Session struct {
	SessionId        string `json:"sessionId,omitempty"`
	Token            string `json:"token,omitempty"`
	Phone            string `json:"phone,omitempty"`
	Email            string `json:"email,omitempty"`
	RegistrationType int    `json:"registrationType"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	CreatedAt        int    `json:"createdAt"`
	UpdatedAt        int    `json:"updatedAt"`
	Salt             []byte `json:"salt,omitempty"`
}
