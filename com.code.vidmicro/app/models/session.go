package models

import (
	"github.com/bytedance/sonic"
)

//This structure mainly handle the session
/*
registrationType variable can have following values
1- System Registration
2- Google Registration
*/
type Session struct {
	SessionId    string `json:"sessionId,omitempty"`
	UserId       int64  `json:"userId,omitempty"`
	Username     string `json:"username,omitempty"`
	Token        string `json:"token,omitempty"`
	Name         string `json:"name,omitempty"`
	Email        string `json:"email,omitempty"`
	Password     string `json:"password,omitempty"`
	AvatarUrl    string `json:"avatar_url,omitempty"`
	IsVerified   bool   `json:"is_verified,omitempty"`
	BlackListed  bool   `json:"black_listed,omitempty"`
	Salt         []byte `json:"salt,omitempty"`
	Role         int    `json:"role,omitempty"`
	RoleName     string `json:"roleName,omitempty"`
	CookieValue  []byte `json:"cookieKey,omitempty"`
	LastActivity int64  `json:"lastActivity,omitempty"`
	CreatedAt    int64  `json:"createdAt,omitempty"`
	ExpiringAt   int64  `json:"expiringAt,omitempty"`
}

func (ts *Session) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *Session) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
