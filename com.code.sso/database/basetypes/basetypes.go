package basetypes

type CollectionName string
type DBName string
type DbType int

const (
	MYSQL DbType = 1
	PSQL  DbType = 2
)
