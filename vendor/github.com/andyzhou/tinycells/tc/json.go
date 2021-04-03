package tc

import (
	"encoding/json"
	"log"
	"runtime/debug"
)

/*
 * base json interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type BaseJson struct {
 }

 //construct
func NewBaseJson() *BaseJson {
	this := &BaseJson{}
	return this
}

 //encode self
func (j *BaseJson) EncodeSelf() []byte {
	var result = make([]byte, 0)

	//encode json
	resp, err := json.Marshal(j)
	if err != nil {
		log.Println("BaseJson::EncodeSelf failed, err:", err.Error())
		return result
	}
	result = resp
	return result
}

//encode json data
func (j *BaseJson) Encode(i interface{}) []byte {
	var result = make([]byte, 0)

	//encode json
	resp, err := json.Marshal(i)
	if err != nil {
		log.Println("BaseJson::Encode failed, err:", err.Error())
		return result
	}
	result = resp
	return result
}

//decode json data
func (j *BaseJson) Decode(data []byte, i interface{}) bool {
	if len(data) <= 0 {
		return false
	}
	//try decode json data
	err := json.Unmarshal(data, i)
	if err != nil {
		//log.Println("BaseJson::Decode, decode failed, err:", err.Error())
		log.Printf("BaseJson::Decode, decode failed: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Println("sakura response:", string(data))
		log.Println("BaseJson::Decode, track:", string(debug.Stack()))
		return false
	}
	return true
}

//encode simple kv data
func (j *BaseJson) EncodeSimple(data map[string]interface{}) []byte {
	if data == nil {
		return nil
	}
	//try encode json data
	byte, err := json.Marshal(data)
	if err != nil {
		log.Printf("BaseJson::EncodeSimple, encode failed: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		return nil
	}
	return byte
}

//decode simple kv data
func (j *BaseJson) DecodeSimple(data []byte, kv map[string]interface{}) bool {
	if len(data) <= 0 {
		return false
	}
	//try decode json data
	err := json.Unmarshal(data, &kv)
	if err != nil {
		log.Printf("BaseJson::DecodeSimple, decode failed: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("sakura response: %q", data)
		return false
	}
	return true
}
