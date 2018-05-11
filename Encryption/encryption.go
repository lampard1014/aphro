package Encryption

import (
    "github.com/xxtea/xxtea-go/xxtea"
    "encoding/pem"
    "encoding/base64"
    "os"
    "io/ioutil"
    "crypto/rsa"
    "crypto/x509"
    "crypto/rand"
    "crypto/sha256"
    "strings"
    "github.com/lampard1014/aphro/Gateway/error"
)

/*
base64Encode
base64Decode 
*/
func  Base64Encode(rawValue []byte) (encodedStr string, err error) {
    str := base64.StdEncoding.EncodeToString(rawValue)
    return str,nil
}

func Base64Decode(decodedStr string) ([]byte, error) {
    decodeBytes , err := base64.StdEncoding.DecodeString(decodedStr)
    return decodeBytes, err
}

func XxteaEncryption(key string, rawValue string) ([]byte, error) {
    encrypt_data := xxtea.Encrypt([]byte(rawValue),[]byte(key))
    return encrypt_data,nil
}

func XxteaDecryption(key string,encryptedStr []byte) (string, error) {
    decrypt_data := xxtea.Decrypt([]byte(encryptedStr),[]byte(key))
    return string(decrypt_data),nil
}

func RsaEncryption(rawValue []byte) ([]byte, error) {
   encryptedData,err := RsaEncrypt(rawValue)
   return encryptedData,err;
}

func RsaDecryption(encryptedStr []byte) ([]byte, error) {
    decryptedData,err := RsaDecrypt(encryptedStr)
    return decryptedData,err
}

var pemMap = map[string]string{"public": "./rsa/public.pem", "private": "./rsa/private.pem"}

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


/*Upper Biz*/
func PswEncryption(psw string) (encryptionPsw string) {
    h := sha256.New()
    h.Write([]byte(psw))
    encryptionPsw = base64.URLEncoding.EncodeToString(h.Sum(nil))
    return encryptionPsw
}


func ParseUsernameAndPsw(key string)(username string ,psw string, err error) {
    base64DecodeRes, base64DecodeErr := Base64Decode(key)
    if base64DecodeErr == nil {
        rawData, RSADecryptionErr := RsaDecryption(base64DecodeRes)
        if RSADecryptionErr == nil {
            usernameAndPsw := string(rawData)
            tmpSplit := strings.Split(usernameAndPsw,"@|@")
            if 2 == len(tmpSplit) {
                username = tmpSplit[0]
                psw = tmpSplit[1]
            } else {
                err = AphroError.New(AphroError.BizError,"拆分用户名密码错误")
            }
        } else {
            err = RSADecryptionErr
        }
    } else {
        err = base64DecodeErr
    }
    return username,psw,err
}
