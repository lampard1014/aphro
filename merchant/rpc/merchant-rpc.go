package  main

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	merchantServicePB "github.com/lampard1014/aphro/merchant/pb"
    encryptionServicePB "github.com/lampard1014/aphro/encryption/encryption-pb"
    sessionServicePB "github.com/lampard1014/aphro/sesssion/pb"
    redisPb "github.com/lampard1014/aphro/redis/pb"
    "os"
    "fmt"
    "strings"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"

)

const (
	port  = ":10089"
    redisRPCAddress = "127.0.0.1:10101"
    sessionRPCAddress = "127.0.0.1:10101"
    encyptRPCAddress = "127.0.0.1:10088"
    mysqlDSN = "root:@tcp(127.0.0.1:3306)/iris_db"
)

type merchantService struct{}


func parseUsernameAndPsw(key string)(username string ,psw string, err error) {

    conn,err := grpc.Dial(encyptRPCAddress,grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    c := encryptionServicePB.NewEncryptionService(conn)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    res, err := c.Base64Decode(ctx,&encryptionServicePB.EncryptionBase64DecodeRequest:{DecodedStr:key})
    if err != nil {
        panic(err)
    }
    rawData, err := c.RsaDecryption(ctx, &encryptionServicePB.DecryptionRSARequest{EncryptedStr:res.RawValue})
    if err != nil {
        panic(err)
    }

    usernameAndPsw := string(rawData.RawValue)
    tmpSplit := strings.Split(usernameAndPsw,"@|@")

    var username string = ""
    var psw string = ""

    if 2 == len(tmpSplit) {
        username = tmpSplit[0]
        psw = tmpSplit[1]
    } else {
        err := errors.New("拆分用户名密码错误")
    }

    return username,psw,err
} 

func (s *merchantService) MerchantOpen(ctx context.Context, in *merchantServicePB.MerchantOpenRequest) (*merchantServicePB.MerchantOpenResponse, error) {

    merchantName := in.Name
    cellphone := in.Cellphone
    address := in.Address
    paymentBit := in.PaymentBit

    db, err := sql.Open("mysql", mysqlDSN)
    defer db.Close()
    // Open doesn't open a connection. Validate DSN data:
    err = db.Ping()
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }

    stmtIns, err := db.Prepare("INSERT INTO merchant (`merchant_name`,`merchant_address`,`payment_type`,`cellphone`) VALUES( ?, ?, ?, ?)") // ? = placeholder
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }
    defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

    _, err = stmtIns.Exec(merchantName, address, paymentBit, cellphone) 
    if err != nil {
        panic(err.Error())
    }

    return &merchantServicePB.MerchantOpenResponse{Successed:err == nil}
}


func (s *merchantService) MerchantRegister(ctx context.Context, in *merchantServicePB.MerchantRegisterRequest) (*merchantServicePB.MerchantRegisterResponse, error) {
    cellphone, psw, err:= parseUsernameAndPsw(in.Key)
    name := in.name
    role := in.Role
    verifyCode := in.VerifyCode
    merchantID := in.MerchantID

    operatorSuccess := false

    if err != nil {
        panic(err)
    }

    //验证码检查
    conn,err := grpc.Dial(redisRPCAddress,grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    c := redisPb.NewRedisServiceClient(conn)
    ctxRPC, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    checkKey := "_verify_sms_"+ cellphone + "@1" 

    queryRes, err1 := c.Query(ctxRPC, &redisPb.QueryRedisRequest{Key: checkKey})

    hasError := err != nil || err1 != nil

    verifyCodeBingo := false

    if !hasError && queryRes!= nil && queryRes.Value ==  verifyCode {
        verifyCodeBingo = true
    }
    
    if verifyCodeBingo {
        //插入数据库
        db, err := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        // Open doesn't open a connection. Validate DSN data:
        err = db.Ping()
        if err != nil {
            panic(err.Error()) // proper error handling instead of panic in your app
        }

        stmtIns, err := db.Prepare("INSERT INTO `merchant_account` (`name`,`cellphone`,`psw`,`role`,`merchant_id`)VALUES(?,?,?,?,?)") // ? = placeholder
        if err != nil {
            panic(err.Error()) // proper error handling instead of panic in your app
        }
        defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

        _, err = stmtIns.Exec(name, cellphone, psw, role,merchantID) 
        if err != nil {
            panic(err.Error())
        } else {
            operatorSuccess = true
        }
    }

    return &merchantServicePB.MerchantRegisterResponse{Successed:operatorSuccess}
}

func (s *merchantService) MerchantChangePsw(ctx context.Context, in *MerchantChangePswRequest) (*MerchantChangePswResponse, error) {

}

func (s *merchantService) MerchantLogin(ctx context.Context, in *MerchantLoginRequest) (*MerchantLoginResponse, error) {
    cellphone, psw, err:= parseUsernameAndPsw(in.Key)
    name := in.name
    tokenRequest := in.TokenRequest
    merchantID := in.MerchantID

    operatorSuccess := false

    if err != nil {
        panic(err)
    }
    //查询数据库
    var (
        uid uint32
        name string
        role int
        merchantID uint32
    )

    db, err := sql.Open("mysql", mysqlDSN)
    defer db.Close()
    // Open doesn't open a connection. Validate DSN data:
    err = db.Ping()
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }
    err := db.QueryRow("SELECT id as uid, name ,role,merchant_id FROM merchant_account WHERE cellphone = ? AND psw = ? LIMIT 1", cellphone, psw).
    Scan(&uid,&name,&role,&merchantID)
    if err != nil {
        log.Fatal(err)
    } else if err == sql.ErrNoRows{

    } else {
        //生成令牌

        conn,err := grpc.Dial(sessionRPCAddress,grpc.WithInsecure())
        if err != nil {
            panic(err)
        }
        defer conn.Close()
        c := sessionServicePB.NewSessionServiceClient(conn)
        ctxRPC, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()

        if tokenRequest {
            
            connRSA,errRSA := grpc.Dial(encyptRPCAddress,grpc.WithInsecure())
            if errRSA != nil {
                panic(err)
            }
            defer connRSA.Close()
            cRSA := encryptionServicePB.NewEncryptionService(conn)
            ctxRSA, cancelRSA := context.WithTimeout(context.Background(), time.Second)
            defer cancel()

            res, err := c.Base64Decode(ctx,&encryptionServicePB.EncryptionBase64DecodeRequest:{DecodedStr:tokenRequest})
            if err != nil {
                panic(err)
            }
            rawData, err := c.RsaDecryption(ctx, &encryptionServicePB.DecryptionRSARequest{EncryptedStr:res.RawValue})
            if err != nil {
                panic(err)
            }
            //得到xxtea key





        }

        checkKey := "_usertoken_"+ cellphone + "@" 

        queryRes, err1 := c.CreateSessionToken(ctxRPC, &sessionServicePB.SessionTokenCreateRequest{SessionTokenRequestStr: checkKey})

        hasError := err != nil || err1 != nil

        verifyCodeBingo := false

  


    }


    return &merchantServicePB.MerchantLoginResponse{Successed:operatorSuccess}
}

func (s *merchantService) MerchantInfo(ctx context.Context, in *MerchantInfoRequest) (*MerchantInfoResponse, error) {

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
