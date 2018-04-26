package  main

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    "github.com/xxtea/xxtea-go/xxtea"
	pb "github.com/lampard1014/aphro/encryption/encryption-pb"
    "encoding/pem"
    "encoding/base64"
    "os"
    "io/ioutil"
    "crypto/rsa"
    "crypto/x509"
    "crypto/rand"
    "fmt"
)

const (
	port  = ":10087"
)

type encryptionService struct{}

/*
base64Encode
base64Decode 
*/

func (s *encryptionService) Base64Encode(ctx context.Context, in *pb.EncryptionBase64EncodeRequest) (*pb.EncryptionBase64EncodeResponse, error) {
    str := base64.StdEncoding.EncodeToString(in.RawValue)
    return &pb.EncryptionBase64EncodeResponse{EncodedStr:str},nil
}

func (s *encryptionService) Base64Decode(ctx context.Context, in *pb.EncryptionBase64DecodeRequest) (*pb.EncryptionBase64DecodeResponse, error) {
    decodeBytes , err := base64.StdEncoding.DecodeString(in.DecodedStr)
    return &pb.EncryptionBase64DecodeResponse{RawValue:decodeBytes},err
}

func (s *encryptionService) XxteaEncryption(ctx context.Context, in *pb.EncryptionXXTEARequest) (*pb.EncryptionXXTEAResponse, error) {
    encrypt_data := xxtea.Encrypt([]byte(in.RawValue),[]byte(in.Key))
    return &pb.EncryptionXXTEAResponse{Key:in.Key,EncryptedStr:encrypt_data},nil
}

func (s *encryptionService) XxteaDecryption(ctx context.Context, in *pb.DecryptionXXTEARequest) (*pb.DecryptionXXTEAResponse, error) {
    decrypt_data := xxtea.Decrypt([]byte(in.EncryptedStr),[]byte(in.Key))
    return &pb.DecryptionXXTEAResponse{Key:in.Key,RawValue:string(decrypt_data)},nil
}

func (s *encryptionService) RsaEncryption(ctx context.Context, in *pb.EncryptionRSARequest) (*pb.EncryptionRSAResponse, error) {
   encrypedData,err := RsaEncrypt(in.RawValue)
   return &pb.EncryptionRSAResponse{EncryptedStr:encrypedData},err;
}

func (s *encryptionService) RsaDecryption(ctx context.Context, in *pb.DecryptionRSARequest) (*pb.DecryptionRSAResponse, error) {
    decryptdData,err := RsaDecrypt(in.EncryptedStr)
    return &pb.DecryptionRSAResponse{RawValue:decryptdData},err
}

var pemMap = map[string]string{"public": "../rsa/public.pem", "private": "../rsa/private.pem"}

func GetBlockFromPem(key string) []byte {
    path := pemMap[key]
    fi, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    defer fi.Close()
    fd, err := ioutil.ReadAll(fi) //读取文件内容

     pemKey := []byte(string(fd))
    block, _ := pem.Decode(pemKey)
    if block == nil {
        panic(key + " key error!")
    }
    return block.Bytes
}

// 加密
func RsaEncrypt(origData []byte) ([]byte ,error){
    publicPem := GetBlockFromPem("public") //获取公钥pem的block
    pubInterface, err := x509.ParsePKIXPublicKey(publicPem) //解析公钥
    if err != nil {
        panic(err)
    }
    pub := pubInterface.(*rsa.PublicKey)
    encypt, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
    if err != nil {
        panic(err)
    }
    return encypt,err
}

// 解密
func RsaDecrypt(encypt []byte) ([]byte ,error){
    privatePem := GetBlockFromPem("private")
    priv, err := x509.ParsePKCS1PrivateKey(privatePem) //解析私钥
    if err != nil {
        panic(err)
    }
    decypt, err := rsa.DecryptPKCS1v15(rand.Reader, priv, encypt)

    if err != nil {
        panic(err)
    }
    return decypt,err
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
    pb.RegisterEncryptionServiceServer(s, new(encryptionService))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }
}
