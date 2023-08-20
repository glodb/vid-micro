package clientmodels

import "com.code.sso/com.code.sso/app/models"

type ClientUser struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    models.User `json:"data"`
}
