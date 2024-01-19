package models

import "encoding/json"

//This structure mainly handle the session
/*
registrationType variable can have following values
1- System Registration
2- Google Registration
*/
type Session struct {
	UserId           string `json:"userId,omitempty"`
	SessionId        string `json:"sessionId,omitempty"`
	Token            string `json:"token,omitempty"`
	Phone            string `json:"phone,omitempty"`
	Email            string `json:"email,omitempty"`
	RegistrationType int    `json:"registrationType"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	LastActivity     int64  `json:"lastActivity"`
	Role             int    `json:"role"`
	CreatedAt        int    `json:"createdAt"`
	UpdatedAt        int    `json:"updatedAt"`
	Salt             []byte `json:"salt,omitempty"`
}

func (ts *Session) EncodeRedisData() []byte {
	buf, _ := json.Marshal(ts)
	return buf
}

func (ts *Session) DecodeRedisData(data []byte) {
	json.Unmarshal(data, &ts)
}
