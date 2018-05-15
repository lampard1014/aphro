package Response

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"log"
	"golang.org/x/net/context"
	"github.com/lampard1014/aphro/CommonBiz/Response/PB"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/lampard1014/aphro/Gateway/error"
	"strconv"
)


const (
	//NoError
	NoError  = iota + 100
	//验签错误
	AuthError
	//业务逻辑错误
	BizError
	//  common biz error //
	//session过期
	SessionExpired

)

func NewCommonBizResponse(code int64, message string,resultMsg proto.Message )(*Aphro_CommonBiz.Response,error) {

	if resultMsg == nil {
		return &Aphro_CommonBiz.Response{code,message,nil},nil
	} else {
		any, err := MarshalAny(resultMsg)
		r := &Aphro_CommonBiz.Response{code,message,any}
		return r,err
	}
}

func NewCommonBizResponseWithCodeWithError(code int64, err error,resultMsg proto.Message )(*Aphro_CommonBiz.Response,error) {
	var errMsg string = ""
	if err != nil {
		errMsg = err.Error()
	}
	return NewCommonBizResponse(code,errMsg,resultMsg)
}

func NewCommonBizResponseWithError(err error,resultMsg proto.Message )(*Aphro_CommonBiz.Response,error) {

	var msg  string
	var code int64 = int64(BizError)
	if d,ok := err.(*AphroError.CustomGRPCError); ok {
		code = int64(d.Code)
		msg = d.Message
	} else {

		if err != nil {
			msg = err.Error()
		}
		if tmpCode, ok := strconv.Atoi(msg); ok == nil  && tmpCode != 0 {
			code = int64(tmpCode)
		}
	}
	return NewCommonBizResponse(code,msg,resultMsg)

}

func MarshalAny(protoMsg proto.Message)(*any.Any, error) {
	any, err := ptypes.MarshalAny(protoMsg)
	return any,err
}

func UnmarshalAny(any *any.Any, pb proto.Message)(error) {
	return ptypes.UnmarshalAny(any,pb)
}


func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("before handling. Info: %+v", info)
	resp, err := handler(ctx, req)

	v,_ := resp.(proto.Message)

	var code int64 = 0
	if err != nil{
		code = 1
	}
	var message string
	if err != nil {
		message = err.Error()
	}

	x,err := NewCommonBizResponse(code, message, v)

	log.Printf("after handling. resp: %+v", x)
	return x, err
}
// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Printf("before handling. Info: %+v", info)
	err := handler(srv, ss)
	log.Printf("after handling. err: %v", err)
	return err
}