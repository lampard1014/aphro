package Error
import (

	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
 )

//type CustomCode codes.Code

const (
	//NoError
	NoError  = 0
	//验证错误
	AuthError  = 100 +iota
	//业务逻辑错误
	BizError
)

type AphroError interface {
	Code()int
	Message()string
}

type CustomError struct{
	code int
	message string
}

//实现 AphroError interface
func (e *CustomError) Code()int {
	return e.code
}

func (e *CustomError) Message()string {
	return e.message
}

//func (e *CustomError) SetCode(c int) {
//	e.code = c
//}
//
//func (e *CustomError) setMessage(m string) {
//	e.message = m
//}

//实现error 接口
func (e *CustomError) Error() string {
	return e.message
}
//GRPC
func (e *CustomError) GRPCStatus() *status.Status {
	var c codes.Code = codes.Code(BizError)
	_code := e.Code()
	if _code != 0 {
		c = codes.Code(_code)
	}
	return status.New(c,e.Message())
}

//自定义new方法
func NewCustomError(c int, msg string) *CustomError {
	e := &CustomError{}
	e.message = msg
	e.code = c
	return e
}

