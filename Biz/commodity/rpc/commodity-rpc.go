package main

import (
    "github.com/lampard1014/aphro/Biz/commodity/CommdityServiceIMP"
    "github.com/lampard1014/aphro/Biz/commodity/pb"
    "fmt"
    "net"
    "log"
    "google.golang.org/grpc"
)


func deferFunc() {
   if err := recover(); err != nil {
       fmt.Println("error happend:")
       fmt.Println(err)
   }
}

func main()  {
   defer deferFunc()
   lis, err := net.Listen("tcp", CommdityServiceIMP.RPCPort)
   if err != nil {
       log.Fatal(err)
   }
   s := grpc.NewServer()//opts...)
    Aphro_Commodity_pb.RegisterCommodityServiceServer(s,new(CommdityServiceIMP.CommodityServiceImp))
   err = s.Serve(lis)
   if err != nil {
       log.Fatal(err)
   }

}