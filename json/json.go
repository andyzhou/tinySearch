package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

/*
 * face for json
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
func (j *BaseJson) EncodeSelf() ([]byte, error) {
	//encode json
	resp, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//encode json data
func (j *BaseJson) Encode(i interface{}) ([]byte, error) {
	if i == nil {
		return nil, errors.New("invalid parameter")
	}
	//encode json
	resp, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//decode json data
func (j *BaseJson) Decode(data []byte, i interface{}) error {
	if len(data) <= 0 {
		return errors.New("json data is empty")
	}
	//try decode json data
	err := json.Unmarshal(data, i)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			err = fmt.Errorf("err:%v, syntax error at byte offset %d", err, e.Offset)
		}
		return err
	}
	return nil
}

//encode simple kv data
func (j *BaseJson) EncodeSimple(data map[string]interface{}) ([]byte, error) {
	if data == nil {
		return nil, errors.New("json data is empty")
	}
	//try encode json data
	byte, err := json.Marshal(data)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		return nil, err
	}
	return byte, nil
}

//decode simple kv data
func (j *BaseJson) DecodeSimple(data []byte, kv map[string]interface{}) error {
	if len(data) <= 0 {
		return errors.New("json data is empty")
	}
	//try decode json data
	err := json.Unmarshal(data, &kv)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("sakura response: %q", data)
		return err
	}
	return nil
}
