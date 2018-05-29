package PersistentStore

//CRUD

type IAphroPersistentStore interface {
	Connect()(IAphroPersistentStore)
	Close()(IAphroPersistentStore)
	Reset()(IAphroPersistentStore)
}

type IAphroSQLPersistentStore interface {
	IAphroPersistentStore
	Query(querySQL string, bindsValues...interface{})(IAphroPersistentStoreResult)
	Select(columnsAs map[string]string)(IAphroSQLPersistentStore)
	Insert(entity string,columns []string,values [][]string)(IAphroSQLPersistentStore)
	Update(entity string,columnValues map[string]interface{})(IAphroSQLPersistentStore)
	Delete()(IAphroSQLPersistentStore)
	From(entity string, entityAlias string)(IAphroSQLPersistentStore)

	InnerJoin(entity string,entityAlais string) (IAphroSQLPersistentStore)
	LeftJoin(entity string,entityAlais string) (IAphroSQLPersistentStore)
	RightJoin(entity string,entityAlais string) (IAphroSQLPersistentStore)

	Where(condition interface{}) (IAphroSQLPersistentStore)

	Execute(bindsValues...interface{})(IAphroPersistentStoreResult)
	Limit(s...string)(IAphroSQLPersistentStore)
	OrderBy(orderBySQL string)(IAphroSQLPersistentStore)
}

type IAphroKVPersistentStore interface {
	IAphroPersistentStore

	IsExists(key string)(isExists bool,err error)
	ExpireAt(key string, ttl int64)(success bool, err error)

	QueryTTL(key string)(ttl int64,err error)

	Query(key string)(value string,err error)

	Delete(key string)(success bool,err error)

	Set(key string ,value string ,ttl int64)(success bool,err error)
}

type IAphroPersistentStoreClientConfiguration interface {
	SetOptions (map[string]interface{})
	GetOptions()(map[string]interface{})
}

//client interface
type IAphroPersistentStoreClient interface {
	FetchClient()(interface{})
	FetchConfiguration()(IAphroPersistentStoreClientConfiguration)
	SetConfiguration(IAphroPersistentStoreClientConfiguration)(error)
}

//result interface
type IAphroPersistentStoreResult interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
	FetchRow(dest...interface{}) (error)
	FetchAll(func(...interface {}), ...interface {}) (error)
}

///////////////////////////////////////////////
//Field
///////////////////////////////////////////////
type IAphroPersistentStoreField interface {
	FetchFieldName() (string,error)
	FetchAlias() (string,error)
}

///////////////////////////////////////////////
//Entity
///////////////////////////////////////////////
type IAphroPersistentStoreEntity interface {
	FetchEntityName()(string,error)
	FetchEntityAlais()(string,error)
}




