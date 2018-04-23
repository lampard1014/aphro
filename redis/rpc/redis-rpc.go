package  main

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/lampard1014/aphro/redis/pb"
    "fmt"
    "github.com/go-redis/redis"

)

const (
	port  = ":10101"
)

type redisService struct{}

func (s *redisService) Query(ctx context.Context, in *QueryRedisRequest) (*QueryRedisResponse, error) {
    cli := createClient()
    val, err := cli.Get(in.Key).Result()
    if err != nil {
        panic(err)
    }
    return &QueryRedisResponse{
        Successed:1,
        Key :in.Key,
        Value:val,
        Ttl:
    }, err
}

func (s *redisService) Update(ctx context.Context, in *UpdateRedisRequest) (*UpdateRedisResponse, error) {
    return nil,nil
}

func (s *redisService) Delete(ctx context.Context, in *DeleteRedisRequest) (*DeleteRedisResponse, error) {
    return nil,nil

}

func (s *redisService) Insert(ctx context.Context, in *InsertRedisRequest) (*InsertRedisResponse, error) {
    return nil,nil

}



var redisCli : *Client = nil

func createClient()(*Client) {
    if redisCli == nil {
        redisCli := redis.NewClient(&redis.Options{
            Addr:     "127.0.0.1:6379",
            Password: "", // no password set
            DB:       0,  // use default DB
        })
        pong, err := client.Ping().Result()
        fmt.Println(pong, err)
    }
    return redisCli
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
