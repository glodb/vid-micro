package clientmodels

import "com.code.sso/com.code.sso/app/models"

type ClientSession struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    models.Session `json:"data"`
}
