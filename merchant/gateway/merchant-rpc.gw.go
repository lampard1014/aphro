package main

import (
  "fmt"
  "reflect"
  "io"
  "flag"
  "net/http"
	"github.com/golang/glog"
  "golang.org/x/net/context"
  "github.com/grpc-ecosystem/grpc-gateway/runtime"
  "google.golang.org/grpc"
  gw "github.com/lampard1014/aphro/merchant/pb"
  "github.com/golang/protobuf/proto"
  "google.golang.org/grpc/status"
  "google.golang.org/grpc/codes"
  // spb "google.golang.org/genproto/googleapis/rpc/status"
    "github.com/lampard1014/aphro/gateway/error"
    "encoding/json"
)

var (
  echoEndpoint = flag.String("echo_endpoint", "localhost:10089", "endpoint of YourService")
)

type CustomError struct {
  Message string
  Code int32
  Result interface{}
  // grpcErrCode codes.Code
}

func NewCustomError (m string , c int32, r interface{}) *CustomError {
  return &CustomError{Message : m , Code: c, Result:r}
}


type RetJsonStruct struct {
  Message string `json:"message"`
  Code int32  `json:"code"`
  Result interface{} `json:"result"`
}

//实现message interface
func (r *RetJsonStruct) Reset() {
  r = &RetJsonStruct{}
}

func (r *RetJsonStruct) String() string {
  return proto.CompactTextString(r)
}

func (r *RetJsonStruct)ProtoMessage() {}

func CustomForwardResponseOption(ctx context.Context,w http.ResponseWriter,pm proto.Message) error {

  // r := http.response(w)
  // fmt.Println("asset", r)

  fmt.Println("[--------------]",pm);
  fmt.Println("[--------------]",w);
  // fmt.Println("str ++> ", pm.String() , pm.ProtoMessage())
  // pm.Reset()


  // hi := w.(http.Hijacker)
  // c,rw,e := hi.Hijack()
  // fmt.Println("c,rw,e ",c,rw,e)

  //retJsonStruct := &RetJsonStruct{
  //  Message:"",
  //  Code:int32(0),
  //  Result:pm,//pm.GetData(),
  //}

  // pm = retJsonStruct
  //b, _ := json.Marshal(retJsonStruct)
  //_, err := w.Write(b)
  return nil
}


func CustomErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
  // return Internal when Marshal failed
  const fallback = `{"code": 13, "message": "failed to marshal error message"}`

  w.Header().Del("Trailer")
  w.Header().Set("Content-Type", marshaler.ContentType())


// NewCustomError
    fmt.Println("errr", err)
    // tmp := reflect.Indirect(err)
    fmt.Println("aa",reflect.TypeOf(err))

  // errClass := reflect.Indirect(reflect.ValueOf(err)).Type().Name()
  // fmt.Println("xxxx",errClass)

  s, ok := status.FromError(err)

  if !ok {
    s = status.New(codes.Unknown, err.Error())
  }

  // var wrapJson = `{"code": 13, "message": "failed to marshal error message"}`

  rawCode := s.Code()
  rawMessage := s.Message()
  rawDetails := s.Details()
  var returnBuf []byte = nil

  buf, merr := marshaler.Marshal(s.Proto())

  fmt.Println("AphroError.BizError",AphroError.BizError,codes.Code(AphroError.BizError),s.Code())

  if s.Code() == codes.Code(AphroError.BizError) {

  } else {
    returnBuf = buf
  }

  retJsonStruct := &RetJsonStruct{
    Message:rawMessage,
    Code:int32(rawCode),
    Result:rawDetails,
  }
  b, _ := json.Marshal(retJsonStruct)
  returnBuf = b

  fmt.Println("buf ->", returnBuf)
  if merr != nil {
    // grpclog.Printf("Failed to marshal error message %q: %v", s.Proto(), merr)
    w.WriteHeader(http.StatusInternalServerError)
    if _, err := io.WriteString(w, fallback); err != nil {
      // grpclog.Printf("Failed to write response: %v", err)
    }
    return
  }

  md, ok := runtime.ServerMetadataFromContext(ctx)
  fmt.Println("md", md)
  if !ok {
    // grpclog.Printf("Failed to extract ServerMetadata from context")
  }

  // handleForwardResponseServerMetadata(w, mux, md)
  // handleForwardResponseTrailerHeader(w, md)
  st := runtime.HTTPStatusFromCode(s.Code())
  // fmt.Println("st",st)

  w.WriteHeader(st)
  if _, err := w.Write(returnBuf); err != nil {
    // grpclog.Printf("Failed to write response: %v", err)
  }

  // handleForwardResponseTrailer(w, md)
}



func run() error {
  ctx := context.Background()
  ctx, cancel := context.WithCancel(ctx)
  defer cancel()

  // mux := runtime.NewServeMux()
  mux := runtime.NewServeMux(
    runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
    runtime.WithProtoErrorHandler(CustomErrorHandler),
    runtime.WithForwardResponseOption(CustomForwardResponseOption))

  opts := []grpc.DialOption{grpc.WithInsecure()}

  err := gw.RegisterMerchantServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint,opts)
    if err != nil {
        return  err
    }

  return http.ListenAndServe(":8089", mux)
}

func main() {
  flag.Parse()
  defer glog.Flush()

  if err := run(); err != nil {
    glog.Fatal(err)
  }
}
