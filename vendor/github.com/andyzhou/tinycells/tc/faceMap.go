package tc

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

/**
 * simple dynamic call instance method
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 *
 * for example:
 	face := base.NewFaceMap()
 	players := player.NewPlayers()
	face.Bind("players", players)
	result, err := face.Call("players", "Test", "this is string test")
	fmt.Println(result[0].String()) //print string type result
 */

 //internal macro variables
 const (
 	MaxInParams = 10
 )

//face map info
type FaceMap struct {
	//faceMap map[string]reflect.Value `face:Value`
	faceMap sync.Map
	callResult []reflect.Value `call reflect value`
	face reflect.Value `face reflect value`
	isOk bool
	tips string
	inParam []reflect.Value
	params int
	callErr error
	i int
	para interface{}
	sync.RWMutex `data locker`
}

//construct
func NewFaceMap() *FaceMap {
	this := &FaceMap{
		faceMap: sync.Map{}, //keep map opt safety
		callResult:make([]reflect.Value, 0),
		inParam:make([]reflect.Value, MaxInParams),
	}
	return this
}

///////
//API
///////

//get face instance
func (f *FaceMap) GetFace(name string) interface{} {
	//face, ok := f.faceMap[name]
	face, ok := f.faceMap.Load(name)
	if !ok {
		return nil
	}
	return face
}

//bind face with name
func (f *FaceMap) Bind(name string, face interface{}) bool {
	if name == "" || face == nil {
		return false
	}
	//check is exists or not
	v := f.GetFace(name)
	if v != nil {
		return true
	}
	f.faceMap.Store(name, reflect.ValueOf(face))
	return true
}

//call method on all faces
func (f *FaceMap) Cast(method string, params ...interface{}) bool {
	var (
		para interface{}
		inParam = make([]reflect.Value, MaxInParams)
		paramNum = 0
		i = 0
	)
	if method == "" {
		inParam = []reflect.Value{}
		return false
	}
	//init parameters
	paramNum = len(params)
	for i, para = range params {
		if i >= MaxInParams {
			break
		}
		inParam[i] = reflect.ValueOf(para)
	}
	//call method on each face
	subFunc := func(key interface{}, face interface{}) bool {
		face2, ok := face.(reflect.Value)
		if ok {
			face2.MethodByName(method).Call(inParam[0:paramNum])
			return true
		}else{
			return false
		}
	}
	f.faceMap.Range(subFunc)

	//reset in param slice
	inParam = []reflect.Value{}
	return true
}

//dynamic call method with parameters support
func (f *FaceMap) Call(name string, method string, params ...interface{}) ([]reflect.Value, error) {
	var (
		tips string
	)
	//check instance
	face, isOk := f.faceMap.Load(name)
	if !isOk {
		tips = fmt.Sprintf("No face instance for name %s", name)
		return nil, errors.New(tips)
	}

	subFace, ok := face.(reflect.Value)
	if !ok {
		tips = fmt.Sprintf("Invalid face instance for name %s", name)
		return nil, errors.New(tips)
	}

	//init parameters
	inParam := make([]reflect.Value, 0)
	totalParas := 0
	//f.params = len(params)
	for _, para := range params {
		if totalParas >= MaxInParams {
			break
		}
		inParam = append(inParam, reflect.ValueOf(para))
		totalParas++
	}

	//dynamic call method with parameter
	callResult := subFace.MethodByName(method).Call(inParam[0:totalParas])

	return callResult, nil
}

