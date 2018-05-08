package main

import (
    "log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    "time"
    "fmt"
    "strings"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    commodityServicePB "github.com/lampard1014/aphro/commodity/pb"
    sessionServicePB "github.com/lampard1014/aphro/session/pb"
    "github.com/lampard1014/aphro/gateway/error"
   
)

const (
    port = ":10085"
    redisRPCAddress = "192.168.140.23:10101"
    sessionRPCAddress = "127.0.0.1:10088"
    encyptRPCAddress = "127.0.0.1:10087"
    mysqlDSN = "root:123456@tcp(192.168.140.23:3306)/iris_db"
)


func fetchSessionTokenValue(sessionToken string) (uid string, merchantID string, err error) {
    var returnErr error = nil
    sessionRPCConn,sessionRPCErr := grpc.Dial(sessionRPCAddress,grpc.WithInsecure())
    if sessionRPCErr == nil {
        sessionRPCConnClient := sessionServicePB.NewSessionServiceClient(sessionRPCConn)

        sessionRPCConnClientCtx, sessionRPCConnClientCancel := context.WithTimeout(context.Background(), time.Second)
        defer sessionRPCConnClientCancel()
        sessionTokenQueryResponse, sessionTokenQueryResponseErr := sessionRPCConnClient.QuerySessionToken(sessionRPCConnClientCtx, &sessionServicePB.SessionTokenQueryRequest{SessionToken:sessionToken})
        if sessionTokenQueryResponseErr == nil && sessionTokenQueryResponse != nil {
            sessionTokenValue := sessionTokenQueryResponse.SessionToken
            splitValue := strings.Split(sessionTokenValue, "#")
            uidAndMerchantID := strings.Split(splitValue[0],"@")
            uid = uidAndMerchantID[0]
            merchantID = uidAndMerchantID[1]
        } else {
            returnErr = AphroError.New(AphroError.BizError,"session 过期 请重新登录")
        }
    } else {
        returnErr = sessionRPCErr
    }
    defer sessionRPCConn.Close()
    return uid,merchantID,returnErr
}

type  commodityService struct{}

func (s *commodityService) CommodityCreate(ctx context.Context, in *commodityServicePB.CommodityCreateRequest) (*commodityServicePB.CommodityCreateResponse, error) {

     
    commodityInfo := in.Good;
    var returnError error = nil
    //验证token的合法性
    uid, merchantID, sessionTokenError := fetchSessionTokenValue(in.Token)

    log.Println(uid, merchantID)

    if sessionTokenError == nil {

        db, dbOpenErr := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        dbOpenErr = db.Ping()
        if dbOpenErr == nil {
            stmtIns, stmtInsErr := db.Prepare("INSERT INTO commodity (`commodity_name`,`commodity_price`,`commodity_count`) VALUES( ?, ?, ?, ?)")
            if stmtInsErr == nil {
                insertResult, insertErr := stmtIns.Exec(commodityInfo.Name, commodityInfo.Price, commodityInfo.Count)
                if insertErr == nil {
                      affectedRow, affectedRowErr := insertResult.RowsAffected() 
                      if affectedRow != 1 || affectedRowErr == nil {
                          returnError = affectedRowErr
                      }
                    
                } else {
                    returnError = insertErr
                }
            } else {
                returnError = stmtInsErr
            }
            defer stmtIns.Close()
        } else {
            returnError = dbOpenErr
        }
    } else {
        returnError = sessionTokenError
    }
   
   return &commodityServicePB.CommodityCreateResponse{Successed: returnError == nil}, returnError
}

    
func (s *commodityService) CommodityDelete(ctx context.Context, in *commodityServicePB.CommodityDeleteRequest) (*commodityServicePB.CommodityDeleteResponse, error) {

     
    commodityInfo := in.Good;
    var returnError error = nil
    //验证token的合法性
    uid, merchantID, sessionTokenError := fetchSessionTokenValue(in.Token)

    log.Println(uid, merchantID)

    if sessionTokenError == nil {

        db, dbOpenErr := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        dbOpenErr = db.Ping()
        if dbOpenErr == nil {
            stmtIns, stmtInsErr := db.Prepare("delete `commodity` where id = ? limit 1")
            if stmtInsErr == nil {
                deleteResult, deleteErr := stmtIns.Exec(commodityInfo.Id)
                if deleteErr == nil {
                      affectedRow, affectedRowErr := deleteResult.RowsAffected() 
                      if affectedRow != 1 || affectedRowErr == nil {
                          returnError = affectedRowErr
                      }
                    
                } else {
                    returnError = deleteErr
                }
            } else {
                returnError = stmtInsErr
            }
            defer stmtIns.Close()
        } else {
            returnError = dbOpenErr
        }
    } else {
        returnError = sessionTokenError
    }
   
   return &commodityServicePB.CommodityDeleteResponse{Successed: returnError == nil}, returnError
}

    

    
func (s *commodityService) CommodityUpdate(ctx context.Context, in *commodityServicePB.CommodityUpdateRequest) (*commodityServicePB.CommodityUpdateResponse, error) {

    commodityInfo := in.Good;
    var returnError error = nil
    //验证token的合法性
    uid, merchantID, sessionTokenError := fetchSessionTokenValue(in.Token)

    log.Println(uid, merchantID)

    if sessionTokenError == nil {

        db, dbOpenErr := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        dbOpenErr = db.Ping()
        if dbOpenErr == nil {
            stmtIns, stmtInsErr := db.Prepare("update `commodity` set name = ?,price =? , count = ? where id = ? limit 1")
            if stmtInsErr == nil {
                updateResult, updateErr := stmtIns.Exec(commodityInfo.Name, commodityInfo.Price, commodityInfo.Count, commodityInfo.Id)
                if updateErr == nil {
                      affectedRow, affectedRowErr := updateResult.RowsAffected() 
                      if affectedRow != 1 || affectedRowErr == nil {
                          returnError = affectedRowErr
                      }
                    
                } else {
                    returnError = updateErr
                }
            } else {
                returnError = stmtInsErr
            }
            defer stmtIns.Close()
        } else {
            returnError = dbOpenErr
        }
    } else {
        returnError = sessionTokenError
    }
   
   return &commodityServicePB.CommodityUpdateResponse{Successed: returnError == nil}, returnError
}



func deferFunc() {
    if err := recover(); err != nil {
        fmt.Println("error happend:")
        fmt.Println(err)
    }
}

func main()  {
    defer deferFunc()
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatal(err)
    }
    s := grpc.NewServer()//opts...)
    commodityServicePB.RegisterCommodityServiceServer(s, new(commodityService))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }

}