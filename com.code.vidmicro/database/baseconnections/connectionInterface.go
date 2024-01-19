package baseconnections

import "com.code.vidmicro/com.code.vidmicro/database/basetypes"

type ConntectionInterface interface {
	CreateConnection() (ConntectionInterface, error)
	GetDB(basetypes.DbType) interface{}
}
