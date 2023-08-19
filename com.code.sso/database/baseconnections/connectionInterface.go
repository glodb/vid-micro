package baseconnections

import "com.code.sso/com.code.sso/database/basetypes"

type ConntectionInterface interface {
	CreateConnection() (ConntectionInterface, error)
	GetDB(basetypes.DbType) interface{}
}
