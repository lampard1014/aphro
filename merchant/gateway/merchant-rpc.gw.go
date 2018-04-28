package main

import (
  "flag"
  "net/http"
	"github.com/golang/glog"
  "golang.org/x/net/context"
  "github.com/grpc-ecosystem/grpc-gateway/runtime"
  "google.golang.org/grpc"
  gw "github.com/lampard1014/aphro/merchant/pb"
)

var (
  echoEndpoint = flag.String("echo_endpoint", "localhost:10089", "endpoint of YourService")
)


func run() error {
  ctx := context.Background()
  ctx, cancel := context.WithCancel(ctx)
  defer cancel()

  // mux := runtime.NewServeMux()
  mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))

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
