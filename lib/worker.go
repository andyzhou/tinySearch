package lib

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * general concurrency worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter type
//gen and bind obj ticker, only one works
type (
	SonWorker struct {
		workerId int32
		queue *Queue
		sync.RWMutex
	}
)

//face info
type Worker struct {
	//basic
	workerMap map[int32]*SonWorker //workerId -> *SonWorker
	workerIdMap map[string]int32 //dataId -> workerId, for bind obj
	workers int32
	//cb func
	cbForQueueOpt func(interface{})(interface{}, error)
	sync.RWMutex
}

//construct
func NewWorker() *Worker {
	this := &Worker{
		workerMap: map[int32]*SonWorker{},
		workerIdMap: map[string]int32{},
	}
	return this
}

//quit
func (f *Worker) Quit() {
	f.Lock()
	defer f.Unlock()
	for k, v := range f.workerMap {
		v.Quit()
		delete(f.workerMap, k)
	}
	atomic.StoreInt32(&f.workers, 0)
	runtime.GC()
}

//set cb for queue opt, STEP-1-1
//if setup, will open queue
func (f *Worker) SetCBForQueueOpt(cb func(interface{}) (interface{}, error)) {
	//check
	if cb == nil {
		return
	}
	f.cbForQueueOpt = cb

	//sync into running son workers
	f.Lock()
	defer f.Unlock()
	for _, v := range f.workerMap {
		if v.queue == nil {
			v.queue = NewQueue()
		}
		v.queue.SetCallback(cb)
	}
}

//create workers, STEP-2
func (f *Worker) CreateWorkers(num int) error {
	//check
	if num <= 0 {
		return errors.New("invalid parameter")
	}

	//init batch son workers with locker
	f.Lock()
	defer f.Unlock()
	for i := 0; i < num; i++ {
		//gen new worker id
		newWorkerId := atomic.AddInt32(&f.workers, 1)

		//init son worker
		sw := NewSonWorker(newWorkerId)

		//set queue cb
		if f.cbForQueueOpt != nil {
			if sw.queue == nil {
				sw.queue = NewQueue()
			}
			sw.queue.SetCallback(f.cbForQueueOpt)
		}

		//sync into run map
		f.workerMap[newWorkerId] = sw
	}
	return nil
}

//send data to one worker, STEP-3
//objIds used for hash calculate value
func (f *Worker) SendData(
		data interface{},
		dataId string,
		needResponses ...bool,
	) (interface{}, error) {
	//check
	if data == nil {
		return nil, errors.New("invalid parameter")
	}
	if f.workers <= 0 {
		return nil, errors.New("no any workers")
	}

	//get son worker
	sonWorker, err := f.GetTargetWorker(dataId)
	if err != nil || sonWorker == nil {
		return nil, err
	}
	//send data to queue
	resp, subErr := sonWorker.queue.SendData(data, needResponses...)
	return resp, subErr
}

//cast data to all workers
func (f *Worker) CastData(data interface{}) error {
	//check
	if data == nil {
		return errors.New("invalid parameter")
	}
	if f.workers <= 0 {
		return errors.New("no any workers")
	}

	//send data to all workers
	for _, v := range f.workerMap {
		v.queue.SendData(data)
	}
	return nil
}

//get workers
func (f *Worker) GetWorkers() int32 {
	return f.workers
}

//get son worker
//extParas -> dataId(string), needBind(bool)
func (f *Worker) GetTargetWorker(
	extParas ...interface{}) (*SonWorker, error) {
	var (
		targetWorkerId int32
		dataId string
		needBind bool
	)
	//check
	if extParas != nil {
		extParaLen := len(extParas)
		switch extParaLen {
		case 1:
			{
				dataId = fmt.Sprintf("%v", extParas[0])
			}
		case 2:
			{
				dataId = fmt.Sprintf("%v", extParas[0])
				needBind, _ = strconv.ParseBool(fmt.Sprintf("%v", extParas[0]))
			}
		}
	}

	//gen hashed worker id
	f.Lock()
	defer f.Unlock()
	if dataId == "" {
		//hashed by rand
		now := time.Now().UnixNano()
		rand.Seed(now)
		targetWorkerId = int32(rand.Int63n(now) % int64(f.workers)) + 1
	}else{
		//hashed by data id
		dataIdAscii, err := f.GetAsciiValue(dataId)
		if err != nil {
			return nil, err
		}
		if needBind {
			//get from cached map
			v, ok := f.workerIdMap[dataId]
			if !ok || v <= 0 {
				//hashed by data id
				targetWorkerId = int32(rand.Int63n(int64(dataIdAscii)) % int64(f.workers)) + 1
				//sync into cache map
				f.workerIdMap[dataId] = targetWorkerId
			}else{
				targetWorkerId = v
			}
		}else{
			//hashed by data id
			targetWorkerId = int32(rand.Int63n(int64(dataIdAscii)) % int64(f.workers)) + 1
		}
	}

	//get target son worker
	v, ok := f.workerMap[targetWorkerId]
	if ok && v != nil {
		return v, nil
	}
	return nil, errors.New("can't get son worker")
}

func (f *Worker) GetWorker(
	workerId int32) (*SonWorker, error) {
	//check
	if workerId <= 0 {
		return nil, errors.New("invalid parameter")
	}
	f.Lock()
	defer f.Unlock()
	v, ok := f.workerMap[workerId]
	if ok && v != nil {
		return v, nil
	}
	return nil, errors.New("no such worker")
}


//get ascii value
func (f *Worker) GetAsciiValue(
	input string,
	sizes ...int) (int, error) {
	var (
		size int
	)
	//check
	if input == "" {
		return 0, errors.New("invalid input parameter")
	}

	//detect assigned size
	if sizes != nil && len(sizes) > 0 {
		size = sizes[0]
	}
	if size <= 0 {
		size = DefaultAsciiSize
	}
	inputLen := len(input)
	if inputLen < size {
		size = inputLen
	}

	//loop check
	finalVal := 1
	for idx, val := range input {
		if idx >= size {
			break
		}
		factor := 10 << idx
		asciiVal := int(val)
		finalVal += asciiVal * factor
	}
	return finalVal, nil
}

////////////////////
//api for son worker
////////////////////

//construct
func NewSonWorker(id int32) *SonWorker {
	//self init
	this := &SonWorker{
		workerId: id,
	}
	return this
}

//quit
func (f *SonWorker) Quit() {
	if f.queue != nil {
		f.queue.Quit()
	}
}

//send data
func (f *SonWorker) SendData(data interface{}) (interface{}, error) {
	//check
	if data == nil {
		return nil, errors.New("invalid parameter")
	}
	if f.queue == nil {
		return nil, errors.New("inter queue not init")
	}

	//send data to queue
	resp, err := f.queue.SendData(data)
	return resp, err
}