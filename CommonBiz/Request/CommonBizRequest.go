package CommonBiz

import (
	"github.com/golang/protobuf/proto"
	"reflect"
	"fmt"
)



type wkt interface{XXX_WellKnownType() string}

type CBRequestErr int
const (
	_ CBRequestErr = 0
	CBRequestErrEmptyRawRequest = iota + 500
)

func generalErr (CBRequestErr)(error) {
	return nil
}

func FilterRequest(rawRequest proto.Message)(res map[string]interface{}, err error) {

	if rawRequest != nil {
		iv := reflect.ValueOf(rawRequest)
		fmt.Println(iv)
		numField := iv.NumField()

		//var s string = "xxx"
		//rawRequest.ProtoMessage()

		var i int = numField -1
		for ;  i > 0 ;i--  {
			property :=iv.Field(i)
			fmt.Println(property.Interface())
			//if wkt1,ok := property.(wkt); ok {
			//	if wkt1 != nil  {
			//		//res[property.]
			//	}
			//}
		}
	} else {
	}

	return ;

	//if wkt, ok := v.(wkt); ok {
	//
	//}
}