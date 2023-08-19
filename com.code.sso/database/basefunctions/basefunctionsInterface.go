package basefunctions

import "com.code.sso/com.code.sso/database/basetypes"

/*
* flyweight interface to separate
* different types of connections
* with db functionality
 */
type BaseFucntionsInterface interface {
	GetFunctions() BaseFucntionsInterface
	EnsureIndex(basetypes.DBName, basetypes.CollectionName, bool) error
	Add(basetypes.DBName, basetypes.CollectionName, interface{}) error
	FindOne(basetypes.DBName, basetypes.CollectionName, interface{}, interface{}) error
	UpdateOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query interface{}, data interface{}, upsert bool) error
	DeleteOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query interface{}) error
}
