package main

import (
    "github.com/lampard1014/aphro/Biz/order/OrderServiceImp"
    "fmt"
    "net"
    "log"
    "google.golang.org/grpc"
    "github.com/lampard1014/aphro/Biz/order/PB"
)


func deferFunc() {
   if err := recover(); err != nil {
       fmt.Println("error happend:")
       fmt.Println(err)
   }
}

func main()  {
   defer deferFunc()
   lis, err := net.Listen("tcp", OrderServiceIMP.RPCPort)
   if err != nil {
       log.Fatal(err)
   }
   s := grpc.NewServer()//opts...)
    Aphro_Order_PB.RegisterOrderServiceServer(s,new(OrderServiceIMP.OrderServiceImp))
   err = s.Serve(lis)
   if err != nil {
       log.Fatal(err)
   }
}