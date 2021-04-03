package tc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

/*
 * json format config file processor
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//define config struct
type Config struct {
	kv map[string]interface{}
}

//construct
func NewConfig() *Config {
	return &Config{
		kv:make(map[string]interface{}),
	}
}

//gt value as slice
func (c Config) GetConfigAsSlice(key string) []interface{} {
	ret := make([]interface{}, 1)
	v := c.GetConfig(key)
	if v == nil {
		return ret
	}
	if value, ok := v.([]interface{}); ok {
		ret = value
	}
	return ret
}

//get value as map[string]interface{}
func (c Config) GetConfigAsMap(key string) map[string] interface{} {
	ret := make(map[string]interface{})
	v := c.GetConfig(key)
	if v == nil {
		return ret
	}
	if value, ok := v.(map[string]interface{}); ok {
		ret = value
	}
	return ret
}


//get value as bool
func (c Config) GetConfigAsBool(key string) bool {
	v := c.GetConfig(key)
	if v == nil {
		return false
	}
	if value, ok := v.(bool); ok {
		return value
	}
	return false
}

//get value as integer
func (c Config) GetConfigAsInteger(key string) int {
	v := c.GetConfig(key)
	if v == nil {
		return 0
	}

	var ret string
	switch v.(type) {
	case float64:
		ret = fmt.Sprintf("%1.0f", v)
	case float32:
		ret = fmt.Sprintf("%1.0f", v)
	case string:
		ret = fmt.Sprintf("%s", v)
	default:
		ret = fmt.Sprintf("%d", v)
	}
	val, err := strconv.Atoi(ret)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	return val
}

//get value as string
func (c Config) GetConfigAsString(key string) string {
	v := c.GetConfig(key)
	if v == nil {
		return ""
	}
	if value, ok := v.(string); ok {
		return value
	}
	return ""
}

//get single k/v
func (c Config) GetConfig(key string) interface{} {
	//map k/v fetch
	if v, ok := c.kv[key];ok{
		//t := reflect.TypeOf(v)
		//fmt.Println("v:", v, "type:", t)
		return v
	}
	return nil
}

//get all config
func (c Config) GetAllConfigs() map[string]interface{} {
	return c.kv
}

//load config
func (c Config) LoadConfig(fileName string) error {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("Read config file ", fileName, " failed, err:", err.Error())
		return err
	}

	if err := json.Unmarshal(bytes, &c.kv); err != nil {
		log.Println("Unmarshal failed, err:", err.Error())
		log.Println("bytes:", string(bytes))
		return err
	}
	//fmt.Println(c.kv)
	return nil
}
