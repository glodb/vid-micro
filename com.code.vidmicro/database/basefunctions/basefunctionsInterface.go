package basefunctions

import (
	"database/sql"

	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
)

/*
* flyweight interface to separate
* different types of connections
* with db functionality
 */
type BaseFucntionsInterface interface {
	GetFunctions() BaseFucntionsInterface
	EnsureIndex(basetypes.DBName, basetypes.CollectionName, interface{}) error
	Add(dbName basetypes.DBName, collectionName basetypes.CollectionName, data interface{}, scan bool) (int64, error)
	AddMany(basetypes.DBName, basetypes.CollectionName, []interface{}, bool) ([]int64, error)
	Find(dbName basetypes.DBName, collectionName basetypes.CollectionName, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool) (*sql.Rows, error)
	UpdateOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query string, data []interface{}, upsert bool) error
	DeleteOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, condition map[string]interface{}, useOr bool, addParenthesis bool) error
	Paginate(dbName basetypes.DBName, collectionName basetypes.CollectionName, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool, pageSize int, page int) (*sql.Rows, int64, error)
	Count(dbName basetypes.DBName, collectionName basetypes.CollectionName, condition map[string]interface{}, useOr bool) (int64, error)
	RawQuery(dbName basetypes.DBName, collectionName basetypes.CollectionName, query string, data []interface{}) (*sql.Rows, error)
}
