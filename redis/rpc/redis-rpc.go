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
        Successed:err == nil,
        Value:val,
    }, err
}

func (s *redisService) IsExists(ctx context.Context, in *IsExistsRequest) (*IsExistsResponse, error) {
    cli := createClient()
    isSuccess,err := cli.Exists(in.Key).Result()
    if err != nil {
        panic(err)
    }
    return &IsExistsResponse{IsExists:isSuccess == 1},err
}

func (s *redisService) ExpireAt(ctx context.Context, in *ExpireAtRequest) (*ExpireAtResponse, error) {
    cli := createClient()
    isSuccess,err := cli.ExpireAt(in.Key,time.Unix(in.Ttl,0)).Result()
    if err != nil {
        panic(err)
    }
    return &IsExistsResponse{Successed:isSuccess},err
}


func (s *redisService) Delete(ctx context.Context, in *DeleteRedisRequest) (*DeleteRedisResponse, error) {
    cli := createClient()
    isSuccess,err := cli.Del(in.Key).Result()
    if err != nil {
        panic(err)
    }
    return &DeleteRedisResponse{Successed:isSuccess == 1},err
}

func (s *redisService) Set(ctx context.Context, in *SetRedisRequest) (*SetRedisResponse, error) {
    cli := createClient()
    isSuccess,err := cli.Set(in.Key,in.Value,in.Ttl).Result()
    if err != nil {
        panic(err)
    }
    return &SetRedisResponse{Successed:err == nil},err
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
