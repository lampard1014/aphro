package PersistentStore

//CRUD
type IAphroPersistentStore interface {
	Connect()(IAphroPersistentStore)

	Query(querySQL string, bindsValue []interface{})(IAphroPersistentStore)
	Select(columnsAs map[string]string)(IAphroPersistentStore)
	Insert(entity string,columns []string,values [][]interface{})(IAphroPersistentStore)
	Update(entity string,columnValues map[string]interface{})(IAphroPersistentStore)
	Delete()(IAphroPersistentStore)
	From(entity string, entityAlias string)(IAphroPersistentStore)

}


type IAphroPersistentStoreClientConfiguration interface {
	SetOptions (map[string]string)
	GetOptions()(map[string]string)
}

//client interface
type IAphroPersistentStoreClient interface {
	FetchClient()(interface{})
	FetchConfiguration()(IAphroPersistentStoreClientConfiguration)
	SetConfiguration(IAphroPersistentStoreClientConfiguration)(error)
}

//result interface
type IAphroPersistentStoreResult interface {
	LastInsertId() (uint64, error)
	RowsAffected() (int64, error)
	FetchRow(dest ...interface{}) (error)
	FetchAll(dest ...interface{}) (error)
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




