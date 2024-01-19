package baseinterfaces

import (
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
)

type Controller interface {
	basefunctions.BaseFucntionsInterface
	BaseControllerFactory
	basevalidators.ValidatorInterface
	SetBaseFunctions(basefunctions.BaseFucntionsInterface)
	GetCollectionName() basetypes.CollectionName
	DoIndexing() error
	RegisterApis()
	GetDBName() basetypes.DBName
}
