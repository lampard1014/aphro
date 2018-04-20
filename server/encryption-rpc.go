package  main

import (
	"log"
	"net"
	"encoding/base64"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    "github.com/xxtea/xxtea-go/xxtea"
	pb "github.com/lampard1014/aphro/encryption-pb"
)

const (
	port  = ":10087"
)

type encryptionService struct{}

/*
base64Encode
base64Decode 
*/
func (s *encryptionService ) EncryptionWithXXTEA(ctx context.Context, in *pb.EncryptionXXTEAStrRequest) (*pb.EncryptionXXTEAStrResponse, error) {
    encrypt_data := xxtea.Encrypt([]byte(in.Str.RawValue),[]byte(in.XXTEAKey))
	encodedStr := base64.StdEncoding.EncodeToString(encrypt_data)
    return &pb.EncryptionXXTEAStrResponse{XXTEAKey: in.XXTEAKey,Str:&pb.EncryptionStrResponse {EncryptStr:encodedStr}}, nil
}

func (s *encryptionService ) DecryptionWithXXTEA(ctx context.Context, in *pb.DecryptionXXTEAStrRequest) (*pb.DecryptionXXTEAStrResponse, error) {
	decodeStr ,_:= base64.StdEncoding.DecodeString(in.Str.EncryptStr)
    decrypt_data := xxtea.Decrypt([]byte(decodeStr),[]byte(in.XXTEAKey))
    return &pb.DecryptionXXTEAStrResponse{XXTEAKey: in.XXTEAKey,Str:&pb.DecryptionStrResponse{RawValue:string(decrypt_data)}}, nil
}

func main() {
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatal(err)
    }

    s := grpc.NewServer()
    pb.RegisterEncryptionServiceServer(s, new(encryptionService))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }
}
