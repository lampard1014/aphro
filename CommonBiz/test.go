package main

import (
	"github.com/golang/protobuf/proto"
	"reflect"
	"fmt"
)

type Request struct  {
	data map[string]interface{}
	name string
}


func FilterEmptyRequest(rawRequest proto.Message)(res []string, err error) {

	res = map[string]interface{}{}
	if rawRequest != nil {
		iv := reflect.ValueOf(rawRequest)
		it := reflect.TypeOf(rawRequest)
		elem := iv.Elem()
		//fmt.Println(elem)
		numField := elem.NumField()

		var i int = numField -1
		for ;  i >= 0 ;i--  {
			p := elem.Field(i)
			propertyName := it.Elem().Field(i).Name
			data := p.Interface()
			d,ok := data.(interface{ XXX_WellKnownType() string })
			if d == nil {
				//filter
			} else {
				res[propertyName] = data
			}
		}
	} else {
	}

	return ;

}

type testStruct  struct {

	A string
	B int
	//SessionToken *google_protobuf1.StringValue `protobuf:"bytes,1,opt,name=sessionToken" json:"sessionToken,omitempty"`
	C *testStruct
}


func (s *testStruct) Reset() {

}
func (s *testStruct)String() string {
	return "123123"
}
func (s *testStruct) ProtoMessage() {

}

func main() {

	//x := new(google_protobuf1.StringValue);
	//x.Value = "ffff"
	//fmt.Println(x)
	res ,_ := FilterRequest(&testStruct{"12",1,nil})
	x := new(Request)
	x.data = make(map[string]interface{})
	x.data["xxx"] = 1
	fmt.Println(x)
	fmt.Println(res)
}