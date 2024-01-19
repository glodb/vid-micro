package clientmodels

import "com.code.vidmicro/com.code.vidmicro/app/models"

type ClientUser struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    models.User `json:"data"`
}
