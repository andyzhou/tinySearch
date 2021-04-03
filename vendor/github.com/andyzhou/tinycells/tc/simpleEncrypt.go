package tc

import (
	"bytes"
	"encoding/base64"
	"log"
	"runtime/debug"
	"time"
)

/*
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * Simple encrypt algorithm
 */

 //inter macro define
 const (
	SimpleEncryptKeySize = 32
 )

 //simple encrypt info
 type SimpleEncrypt struct {
 	key string `base security key`
 	Utils
 }
 
 //construct
func NewSimpleEncrypt(securityKey string) *SimpleEncrypt {
	//self init
	this := &SimpleEncrypt{
		key:securityKey,
	}
	return this
}

///////
//api
/////

//decrypt string
func (e *SimpleEncrypt) Decrypt(encStr string) string {
	var (
		encryptStr= bytes.NewBuffer(nil)
		tmpByte byte
	)
	if encStr == "" || e.key == "" {
		return encryptStr.String()
	}
	decodeByte, err := base64.StdEncoding.DecodeString(encStr)
	if err != nil {
		log.Println("SimpleEncrypt::Decrypt failed, err:", err.Error())
		log.Println("SimpleEncrypt::Decrypt, key:", e.key)
		log.Println("SimpleEncrypt::Decrypt, encStr:", encStr)
		log.Println("SimpleEncrypt::Decrypt, trace:", string(debug.Stack()))
		return encryptStr.String()
	}
	decodeStr := string(decodeByte)
	bindKey := e.bindKey(decodeStr)
	bindKeyLen := len(bindKey)
	for i := 0; i < bindKeyLen; i++ {
		if i >= (bindKeyLen - 1) {
			continue
		}
		tmpByte = byte(bindKey[i] ^ bindKey[i+1])
		encryptStr.WriteByte(tmpByte)
		i++
	}
	return encryptStr.String()
}


//encrypt string
func (e *SimpleEncrypt) Encrypt(orgStr string) string {
	var (
		encryptStr = bytes.NewBuffer(nil)
		tmpByte uint8
		j int
	)
	if orgStr == "" || e.key == "" {
		return encryptStr.String()
	}
	mixStr := e.genMixString()
	orgStrLen := len(orgStr)
	for i:= 0; i < orgStrLen; i++ {
		if j == SimpleEncryptKeySize {
			j = 0
		}
		tmpByte = byte(orgStr[i]^mixStr[j])
		encryptStr.WriteByte(byte(mixStr[j]))
		encryptStr.WriteByte(tmpByte)
		j++
	}
	bindKey := e.bindKey(encryptStr.String())
	return base64.StdEncoding.EncodeToString([]byte(bindKey))
}

//decrypt string

///////////////
//private func
///////////////

//bind key
func (e *SimpleEncrypt) bindKey(orgStr string) string {
	var (
		tmpByte byte
		buffStr = bytes.NewBuffer(nil)
		i, j int
	)
	entryKey := e.GenMd5(e.key)
	orgStrLen := len(orgStr)
	for i = 0; i < orgStrLen; i++ {
		if j == SimpleEncryptKeySize {
			j = 0
		}
		tmpByte = byte(orgStr[i]^entryKey[j])
		buffStr.WriteByte(tmpByte)
		j++
	}
	return buffStr.String()
}

//generate mix string
func (e *SimpleEncrypt) genMixString() string {
	now := int(time.Now().Unix())
	randStr := e.Seconds2TimeStr(now)
	md5Str := e.GenMd5(randStr)
	return md5Str
}
