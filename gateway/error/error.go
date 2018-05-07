package AphroError
import (
	// "errors"
// 	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"fmt"
 )

type CustomCode codes.Code

const (
	//NoError
	NoError CustomCode = iota + 100
	//验签错误
	AuthError 
	//业务逻辑错误
	BizError
)

type CustomGRPCError struct{
	Code CustomCode
	Message string
}

//实现error 接口
func (e *CustomGRPCError) Error() string {
	return e.Message
}

func (e *CustomGRPCError) GRPCStatus() *status.Status {
	fmt.Println("called!!!!!!!!!!")
	return status.New(codes.Code(BizError),e.Message)
}

//自定义new方法
func New(c CustomCode, msg string) *CustomGRPCError {
	return &CustomGRPCError{Code:c,Message:msg}
}

