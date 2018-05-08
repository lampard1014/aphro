package PersistentStore

//CRUD
type IAphroPersistentStore interface {
	Query()(IAphroPersistentStoreResult)
	Insert()(IAphroPersistentStoreResult)
	Update()(IAphroPersistentStoreResult)
	Delete()(IAphroPersistentStoreResult)
}

//result interface
type IAphroPersistentStoreResult interface {
	FetchErr() error
	FetchRes() interface{}
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

/////abstract /////

///////////////////////////////////////////////
//field的结构体，实现接口IAphroPersistentStoreField
///////////////////////////////////////////////
type APSField struct {
	filedName string
	alias string
}

func (apsField *APSField) FetchFieldName() (string, error) {
	var returnErr error = nil
	if apsField.filedName == "" {
		returnErr = NewPSErrC(NoFieldSpecify)
	}
	return apsField.filedName,returnErr
}

func (apsField *APSField) FetchAlias() (string, error) {
	var returnErr error = nil
	if apsField.alias == "" {
		returnErr = NewPSErrC(NoFieldAliasSpecify)
	}
	return apsField.alias,returnErr
}

type APSEntity struct {
	entityName string
	alias string
}

func (apsEntity *APSEntity) FetchEntityName() (string, error) {
	var returnErr error = nil
	if apsEntity.entityName == "" {
		returnErr = NewPSErrC(NoEntitySpecify)
	}
	return apsEntity.entityName,returnErr
}

func (apsEntity *APSEntity) FetchEntityAlais() (string, error) {
	var returnErr error = nil
	if apsEntity.alias == "" {
		returnErr = NewPSErrC(NoEntityAliasSpecify)
	}
	return apsEntity.alias,returnErr
}
