package basefunctions

import (
	"errors"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
)

type baseFunctions struct {
	dbfunctions map[basetypes.DbType]*BaseFucntionsInterface
}

var instance *baseFunctions
var once sync.Once

// Singleton. Returns a single object of Factory
// This is pure lazy factory, doesnot even create functions class till dbname is specifically passed
// This also part of flyweight design pattern
func GetInstance() *baseFunctions {
	// var instance
	once.Do(func() {
		instance = &baseFunctions{}
		instance.dbfunctions = make(map[basetypes.DbType]*BaseFucntionsInterface)
	})
	return instance
}

func (u *baseFunctions) GetFunctions(dbType basetypes.DbType, dbName basetypes.DBName) (*BaseFucntionsInterface, error) {
	if connection, ok := u.dbfunctions[dbType]; ok {
		return connection, nil
	}
	switch dbType {
	case basetypes.MYSQL:
		{
			connection := MySqlFunctions{}
			functionsInterface := connection.GetFunctions()

			u.dbfunctions[dbType] = &functionsInterface
			return u.dbfunctions[dbType], nil
		}
	case basetypes.PSQL:
		{ //Adding this because ken wants to use framework for IOT
			connection := PSqlFunctions{}
			functionsInterface := connection.GetFunctions()

			u.dbfunctions[dbType] = &functionsInterface
			return u.dbfunctions[dbType], nil
		}
	}
	return nil, errors.New("Not configured for this db")
}
