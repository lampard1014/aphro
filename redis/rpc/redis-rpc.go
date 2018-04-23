package  main

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/lampard1014/aphro/redis/pb"
    "fmt"
)

const (
	port  = ":10101"
)

type redisService struct{}


func (s *redisService) Query(ctx context.Context, in *QueryRedisRequest) (*QueryRedisResponse, error) {

}

func (s *redisService) Update(ctx context.Context, in *UpdateRedisRequest) (*UpdateRedisResponse, error) {

}

func (s *redisService) Delete(ctx context.Context, in *DeleteRedisRequest) (*DeleteRedisResponse, error) {

}

func (s *redisService) Insert(ctx context.Context, in *InsertRedisRequest) (*InsertRedisResponse, error) {

}

func deferFunc() {
    if err := recover(); err != nil {
        fmt.Println("error happend:")
        fmt.Println(err)
    }
}

func main() {
    defer deferFunc() 
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatal(err)
    }

    s := grpc.NewServer()
    RegisterRedisServiceServer(s, new(redisService))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }
}
