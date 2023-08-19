package baseconnections

import (
	"sync"

	"com.code.sso/com.code.sso/database/basetypes"
)

type dbConnections struct {
	dbconnections map[basetypes.DbType]*ConntectionInterface
}

var instance *dbConnections
var once sync.Once

//Singleton. Returns a single object of Factory
//This is pure lazy factory, doesnot even create db connection till dbname is specifically passed
//This also helps in flyweight design pattern
func GetInstance() *dbConnections {
	// var instance
	once.Do(func() {
		instance = &dbConnections{}
		instance.dbconnections = make(map[basetypes.DbType]*ConntectionInterface)
	})
	return instance
}

func (u *dbConnections) GetConnection(dbType basetypes.DbType) ConntectionInterface {
	if connection, ok := u.dbconnections[dbType]; ok {
		return *connection
	}
	switch dbType {
	case basetypes.MYSQL:
		{
			connection := MysqlConnection{}
			mysqlconnector, err := connection.CreateConnection()
			if err != nil {
				return nil
			}
			u.dbconnections[dbType] = &mysqlconnector
			return *u.dbconnections[dbType]
		}
	case basetypes.PSQL:
		{ //Allowing multiple db connections, test didn't ask but just using factory
			return nil
		}
	}
	return nil
}
