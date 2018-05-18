package  main

import (

	"log"
	"net"
	"google.golang.org/grpc"
    //_ "github.com/go-sql-driver/mysql"
    roomServicePB "github.com/lampard1014/aphro/Biz/room/pb"
    "fmt"
    "github.com/lampard1014/aphro/Biz/room/RoomServiceIMP"
)

func deferFunc() {
    if err := recover(); err != nil {
        fmt.Println("error happend:")
        fmt.Println(err)
    }
}

// // auth 验证Token
// func auth(ctx context.Context) error {
//     md, ok := metadata.FromContext(ctx)
//     if !ok {
//         return grpc.Errorf(codes.Unauthenticated, "无Token认证信息")
//     }

//     var (
//         appid  string
//         appkey string
//     )

//     if val, ok := md["appid"]; ok {
//         appid = val[0]
//     }

//     if val, ok := md["appkey"]; ok {
//         appkey = val[0]
//     }

//     if appid != "101010" || appkey != "i am key" {
//         return grpc.Errorf(codes.Unauthenticated, "Token认证信息无效: appid=%s, appkey=%s", appid, appkey)
//     }

//     return nil
// }


func main() {
    defer deferFunc()
    lis, err := net.Listen("tcp", RoomServiceIMP.Port)
    if err != nil {
        log.Fatal(err)
    }

    // var opts []grpc.ServerOption //签名 和 验签

    // // 注册interceptor
    // var interceptor grpc.UnaryServerInterceptor
    // interceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
    //     err = auth(ctx)
    //     if err != nil {
    //         return
    //     }
    //     // 继续处理请求
    //     return handler(ctx, req)
    // }
    // opts = append(opts, grpc.UnaryInterceptor(interceptor))

	//grpc.UnaryInterceptor(Response.UnaryServerInterceptor)
    s := grpc.NewServer()//opts...)
    roomServicePB.RegisterRoomServiceServer(s, new(RoomServiceIMP.RoomServiceImp))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }
}

//func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
//    log.Printf("before handling. Info: %+v", info)
//    resp, err := handler(ctx, req)
//
//    fmt.Println("reflect", reflect.TypeOf(resp))
//
//    CommonBiz.NewCommonBizResponse(0,err.Error(),resp.(*proto.Message))
//
//    log.Printf("after handling. resp: %+v", resp)
//    return resp, err
//}
//// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
//func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
//    log.Printf("before handling. Info: %+v", info)
//    err := handler(srv, ss)
//    log.Printf("after handling. err: %v", err)
//    return err
//}
