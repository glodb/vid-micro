package baseinterfaces

import (
	"com.code.sso/com.code.sso/database/basefunctions"
	"com.code.sso/com.code.sso/database/basetypes"
	"com.code.sso/com.code.sso/httpHandler/basevalidators"
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
