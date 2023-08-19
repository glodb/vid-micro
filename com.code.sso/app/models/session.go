package models

type Session struct {
	SessionId string `json:"sessionId,omitempty"`
	UserId    string `json:"userId,omitempty"`
	Token     string `json:"token,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	LoginType int    `json:"loginType,omitempty"`
}
