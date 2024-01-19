package clientmodels

import "com.code.vidmicro/com.code.vidmicro/app/models"

type ClientSession struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    models.Session `json:"data"`
}
