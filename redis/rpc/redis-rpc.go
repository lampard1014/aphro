package  main

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "github.com/lampard1014/aphro/redis/pb"
    "fmt"
    "time"
    r "github.com/go-redis/redis"

)

const (
	port  = ":10101"
)

type redisService struct{}

func (s *redisService) Query(ctx context.Context, in *pb.QueryRedisRequest) (*pb.QueryRedisResponse, error) {
    cli := createClient()
    val, err := cli.Get(in.Key).Result()
    fmt.Println("getVal",val)
    if val != "" && err != nil {
        panic(err)
    }
    return &pb.QueryRedisResponse{
        Successed:err == nil,
        Value:val,
    }, err
}

func (s *redisService) IsExists(ctx context.Context, in *pb.IsExistsRequest) (*pb.IsExistsResponse, error) {
    cli := createClient()
    isSuccess,err := cli.Exists(in.Key).Result()

    bingo := isSuccess == 1
    // fmt.Println("isExists:",in.Key,isSuccess,err,bingo)
    if err != nil {
        panic(err)
    }
    return &pb.IsExistsResponse{IsExists:bingo},err
}

func (s *redisService) ExpireAt(ctx context.Context, in *pb.ExpireAtRequest) (*pb.ExpireAtResponse, error) {
    cli := createClient()
    isSuccess,err := cli.ExpireAt(in.Key,time.Unix(int64(in.Ttl),0)).Result()
    if err != nil {
        panic(err)
    }
    return &pb.ExpireAtResponse{Successed:isSuccess},err
}

func (s *redisService) QueryTtl(ctx context.Context, in *pb.QueryTtlRedisRequest) (*pb.QueryTtlRedisResponse, error) {
    cli := createClient()
    ttl,err := cli.TTL(in.Key,).Result()
    fmt.Println("err ttl",err)
    if err != nil {
        panic(err)
    }
    return &pb.QueryTtlRedisResponse{Ttl:int64(ttl.Seconds())},err
}

func (s *redisService) Delete(ctx context.Context, in *pb.DeleteRedisRequest) (*pb.DeleteRedisResponse, error) {
    cli := createClient()
    isSuccess,err := cli.Del(in.Key).Result()
    if err != nil {
        panic(err)
    }
    return &pb.DeleteRedisResponse{Successed:isSuccess == 1},err
}

func (s *redisService) Set(ctx context.Context, in *pb.SetRedisRequest) (*pb.SetRedisResponse, error) {
    cli := createClient()
    _,err := cli.Set(in.Key,in.Value,time.Duration(in.Ttl)).Result()
    if err != nil {
        panic(err)
    }
    return &pb.SetRedisResponse{Successed:err == nil},err
}


func createClient()*r.Client {
    redisCli := r.NewClient(&r.Options{
        Addr:     "127.0.0.1:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    pong, err := redisCli.Ping().Result()
    fmt.Println(pong, err,redisCli)
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
    pb.RegisterRedisServiceServer(s, new(redisService))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }
}
