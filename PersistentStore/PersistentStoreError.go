package PersistentStore

type ErrCode int32

const (
	NoError ErrCode = iota
	NoFieldSpecify
	NoEntitySpecify
	NoFieldAliasSpecify
	NoEntityAliasSpecify
	UnknowError
)

var  ErrorMap  = map[ErrCode]string {
	NoError:"no error",
	NoFieldSpecify:"no filed specify....",
	NoEntitySpecify:"no entity specify....",
	NoFieldAliasSpecify:"no field alias specify....",
	NoEntityAliasSpecify:"no entity alias specify....",
	UnknowError:"unknow error ....",
}

type PersistentStoreError struct{
	Code ErrCode
	Message string
}

func  fetchErrMsgByCode(c ErrCode) string {
	return ErrorMap[c]
}

//实现error 接口
func (e *PersistentStoreError) Error() string {
	return e.Message
}
//自定义new方法
func NewPSErr(c ErrCode, msg string) *PersistentStoreError {
	return &PersistentStoreError{Code:c,Message:msg}
}

func NewPSErrC(c ErrCode) *PersistentStoreError {
	return &PersistentStoreError{Code:c,Message:fetchErrMsgByCode(c)}
}
