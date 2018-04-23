package  main

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    "github.com/xxtea/xxtea-go/xxtea"
	"github.com/lampard1014/aphro/session/pb"
    "fmt"
)

const (
	port  = ":10088"
)

type sessionService struct{}

func (s *sessionService) QuerySessionToken(ctx context.Context, in *SessionTokenQueryRequest) (*SessionTokenQueryResponse, error) {
}

func (s *sessionService) CreateSessionToken(ctx context.Context, in *SessionTokenCreateRequest) (*SessionTokenCreateResponse, error) {
}

func (s *sessionService) DeleteSessionToken(ctx context.Context, in *DeleteSessionTokenRequest) (*DeleteSessionTokenResponse, error) {
}

func (s *sessionService) RenewSessionToken(ctx context.Context, in *RenewSessionTokenRequest) (*RenewSessionTokenResponse, error) {
}

func (s *sessionService) IsSessionTokenVailate(ctx context.Context, in *IsSessionTokenVailateRequest) (*IsSessionTokenVailateResponse, error) {
}

func (s *sessionService) MerchantVerifyCode(ctx context.Context, in *MerchantVerifyCodeRequest) (*MerchantVerifyCodeResponse, error) {
}

func (s *sessionService) MerchantSendCode(ctx context.Context, in *MerchantSendCodeRequest) (*MerchantSendCodeResponse, error) {
}


func (s *encryptionService) Base64Encode(ctx context.Context, in *pb.EncryptionBase64EncodeRequest) (*pb.EncryptionBase64EncodeResponse, error) {
    str := base64.StdEncoding.EncodeToString(in.RawValue)
    return &pb.EncryptionBase64EncodeResponse{EncodedStr:str},nil
}

func (s *encryptionService) Base64Decode(ctx context.Context, in *pb.EncryptionBase64DecodeRequest) (*pb.EncryptionBase64DecodeResponse, error) {
    decodeBytes , _ := base64.StdEncoding.DecodeString(in.DecodedStr)
    return &pb.EncryptionBase64DecodeResponse{RawValue:decodeBytes},nil
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
   encrypedData := RsaEncrypt(in.RawValue)
   return &pb.EncryptionRSAResponse{EncryptedStr:encrypedData},nil;
}

func (s *encryptionService) RsaDecryption(ctx context.Context, in *pb.DecryptionRSARequest) (*pb.DecryptionRSAResponse, error) {
    decryptdData := RsaDecrypt(in.EncryptedStr)
    return &pb.DecryptionRSAResponse{RawValue:decryptdData},nil
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
func RsaEncrypt(origData []byte) []byte {
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
    return encypt
}

// 解密
func RsaDecrypt(encypt []byte) []byte {
    privatePem := GetBlockFromPem("private")
    priv, err := x509.ParsePKCS1PrivateKey(privatePem) //解析私钥
    if err != nil {
    panic(err)
    }
    decypt, err := rsa.DecryptPKCS1v15(rand.Reader, priv, encypt)

    if err != nil {
    panic(err)
    }
    return decypt
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
    RegisterSessionServiceServer(s, new(sessionService))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }
}
