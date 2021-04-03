package tc
//
//
//import (
//	"github.com/aws/aws-sdk-go/service/s3"
//	"github.com/aws/aws-sdk-go/aws/credentials"
//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/aws/session"
//	"log"
//	"sync"
//	"math/rand"
//	"net/http"
//	"time"
//	"bytes"
//	"fmt"
//)
//
///*
// * Amazon S3 service
// * @author <AndyZhou>
// * @mail <diudiu8848@163.com>
// * all opt in currency worker pool
// */
//
////s3 data kind
//const (
//	S3DataKindRead = iota
//	S3DataKindSave
//)
//
////internal macro defines
//const (
//	S3SyncChanSize = 64
//	fileTempBuffSie = 1024
//	S3MaxTryTimes = 3
//)
//
////s3 original data
//type S3Data struct {
//	kind int `data kind, read or save`
//	notifyChan chan bool `s3 data send notify chan`
//	receiverChan chan []byte `s3 data receiver chan`
//	filePath string `s3 file full path`
//	dataBuff []byte
//}
//
////S3 client worker
//type S3Worker struct {
//	id int `worker id`
//	bucket string
//	client *s3.S3 `s3 client instance`
//	dataChan chan S3Data
//	closeChan chan bool
//	Utils `anonymous`
//}
//
////S3 config
//type S3Config struct {
//	AccessKey string
//	SecretKey string
//	Region string
//	Token string
//	Bucket string
//	Workers int
//}
//
////S3 service info
//type S3Service struct {
//	conf *S3Config
//	workerIdx int `async worker id`
//	workers map[int]*S3Worker `s3 worker pool`
//	initFinished bool
//	sync.Mutex
//	Utils `anonymous`
//}
//
//
////construct
//func NewS3Service(conf *S3Config) *S3Service {
//	this := &S3Service{
//		conf:conf,
//		initFinished:false,
//		workers:make(map[int]*S3Worker),
//	}
//
//	//init s3 base concurrency workers
//	go this.initBaseWorkers()
//
//	return this
//}
//
/////////////
////API
////////////
//
////quit
//func (s *S3Service) Quit() bool {
//	if len(s.workers) <= 0 {
//		return false
//	}
//	for _, worker := range s.workers {
//		worker.Quit()
//	}
//	return true
//}
//
////remove worker
//func (s *S3Service) RemoveWorker(idx int) bool {
//	if idx <= 0 {
//		return false
//	}
//	s.Lock()
//	delete(s.workers, idx)
//	s.Unlock()
//	return true
//}
//
//////re init worker pool
////func (s *S3Service) ReInitWorkerPool() bool {
////	//check switcher
////	switcher := conf.RunPlumeConf.GetS3Conf().GetSwitcher()
////
////	if !switcher {
////		return false
////	}
////
////	//basic check
////	confWorkers := conf.RunPlumeConf.GetS3Conf().GetMaxWorkers()
////	curWorkers := len(s.workers)
////	diff := 0
////	log.Println("S3Service::ReInitWorkerPool, confWorkers:", confWorkers, ", curWorkers:", curWorkers)
////
////	if confWorkers == curWorkers {
////		//same, do nothing
////		return false
////	}
////
////	if confWorkers > curWorkers {
////		//need make more
////		diff = confWorkers - curWorkers
////		log.Println("S3Service::ReInitWorkerPool, need make ", diff, " workers..")
////		s.initMoreWorkers(diff)
////	}else{
////		//need release some
////		diff = curWorkers - confWorkers
////		log.Println("S3Service::ReInitClientPool, need release ", diff, " workers..")
////		s.releaseWorkers(diff)
////	}
////
////	log.Println("S3Service::ReInitClientPool, now has ", len(s.workers), " workers")
////
////	return true
////}
//
////check async worker init is finished or not
//func (s *S3Service) HasInitFinished() bool {
//	return s.initFinished
//}
//
////read file
//func (s *S3Service) ReadFile(subDir, filePath string) (bool, []byte) {
//	var (
//		worker *S3Worker
//	)
//
//	//basic check
//	if subDir == "" || filePath == "" {
//		return false, nil
//	}
//
//	//init file full path
//	fileFullPath := fmt.Sprintf("%s/%s", subDir, filePath)
//
//	//try get worker with max try times
//	for i := 1; i <= S3MaxTryTimes; i++ {
//		worker = s.getRandomWorker()
//		if worker != nil {
//			break
//		}
//		//sleep for awhile
//		time.Sleep(time.Second/5)
//	}
//	if worker == nil {
//		return false, nil
//	}
//
//	//try catch panic of send data to closed worker chan
//	defer func() {
//		if err := recover(); err != nil {
//			log.Println("S3Service::ReadFile, worker:", worker.id, " panic happend, err:", err)
//		}
//	}()
//
//	//init data receiver chan
//	receiverChan := make(chan []byte)
//
//	//cast data to worker chan
//	s3Data := S3Data{
//		kind:S3DataKindRead,
//		receiverChan:receiverChan,
//		filePath:fileFullPath,
//	}
//	worker.dataChan <- s3Data
//
//	//wait response from worker
//	resp := <- receiverChan
//	close(receiverChan)
//
//	return true, resp
//}
//
////save file
//func (s *S3Service) SaveFile(subDir, filePath string, fileData []byte) bool {
//	var (
//		worker *S3Worker
//	)
//
//	//basic check
//	if subDir == "" || filePath == "" || len(fileData) <= 0 {
//		return false
//	}
//
//	//init file full path
//	fileFullPath := fmt.Sprintf("%s/%s", subDir, filePath)
//
//	//try get worker with max try times
//	for i := 1; i <= S3MaxTryTimes; i++ {
//		worker = s.getRandomWorker()
//		if worker != nil {
//			break
//		}
//		//sleep for awhile
//		time.Sleep(time.Second/5)
//	}
//	if worker == nil {
//		return false
//	}
//
//	//try catch panic of send data to closed worker chan
//	defer func() {
//		if err := recover(); err != nil {
//			log.Println("S3Service::SaveFile, worker:", worker.id, " panic happend, err:", err)
//		}
//	}()
//
//	//init notify chan
//	notifyChan := make(chan bool)
//
//	//cast data to worker chan
//	s3Data := S3Data{
//		kind:S3DataKindSave,
//		notifyChan:notifyChan,
//		filePath:fileFullPath,
//		dataBuff:fileData,
//	}
//	worker.dataChan <- s3Data
//
//	//wait response from worker
//	resp := <- notifyChan
//	close(notifyChan)
//	log.Println("S3Service::SaveFile, fileFullPath:", fileFullPath, ", resp:", resp)
//
//	return true
//}
//
////init internal server rpc clients
//func (s *S3Service) InitRpcClients() bool {
//	return true
//}
//
//
////////////////////
////API for worker
///////////////////
//
////worker quit
//func (w *S3Worker) Quit() {
//	w.closeChan <- true
//	time.Sleep(time.Second/10)
//}
//
////////////////////////////////
////private func for s3 service
///////////////////////////////
//
////get rand worker
//func (s *S3Service) getRandomWorker() *S3Worker {
//	totalWorkers := len(s.workers)
//	randNum := s.getRandomVal(totalWorkers)
//	worker, ok := s.workers[randNum]
//	log.Println("S3Service::getRandomWorker, randNum:", randNum)
//	if !ok {
//		//try init new
//		worker = s.initWorker(randNum)
//		return worker
//	}
//	return worker
//}
//
////get rand number
//func (s *S3Service) getRandomVal(maxVal int) int {
//	randSand := rand.NewSource(time.Now().UnixNano())
//	r := rand.New(randSand)
//	return r.Intn(maxVal)
//}
//
////release some s3 workers from pool
//func (s *S3Service) releaseWorkers(num int) bool {
//	var (
//		worker *S3Worker
//		ok bool
//	)
//	if num <= 0 {
//		return false
//	}
//	idxStart := len(s.workers)
//	idxEnd := idxStart - num
//	for i := idxStart; i > idxEnd; i-- {
//		//close need released worker
//		worker, ok = s.workers[i]
//		if ok {
//			worker.Quit()
//		}
//	}
//	return true
//}
//
////init more s3 worker pool
//func (s *S3Service) initMoreWorkers(num int) bool {
//	if num <= 0 {
//		return false
//	}
//	idxStart := len(s.workers)
//	idxEnd := idxStart + num
//	for i := idxStart; i <= idxEnd; i++ {
//		//init single worker
//		s.initWorker(i)
//	}
//	return true
//}
//
//
////init batch s3 base worker pool
//func (s *S3Service) initBaseWorkers() bool {
//	maxWorkers := s.conf.Workers
//	for i := 1; i <= maxWorkers; i++ {
//		//init single worker
//		s.initWorker(i)
//	}
//	return true
//}
//
////create single worker
//func (s *S3Service) initWorker(idx int) *S3Worker {
//	//init s3 client
//	s3Client := s.initS3Client()
//	if s3Client == nil {
//		return nil
//	}
//
//	//get bucket
//	bucket := s.conf.Bucket
//
//	//init worker
//	worker := &S3Worker{
//		id:idx,
//		bucket:bucket,
//		client:s3Client,
//		dataChan:make(chan S3Data, S3SyncChanSize),
//		closeChan:make(chan bool),
//	}
//
//	//spawn worker main process
//	go worker.runMainProcess()
//
//	//sync into worker map
//	s.Lock()
//	s.workers[idx] = worker
//	s.Unlock()
//
//	return worker
//}
//
////init single s3 client
//func (s *S3Service) initS3Client() *s3.S3 {
//	//get s3 key
//	accessKey := s.conf.AccessKey
//	secretKey := s.conf.SecretKey
//	region := s.conf.Region
//	token := s.conf.Token
//
//	//init credential
//	cred := credentials.NewStaticCredentials(accessKey, secretKey, token)
//	_, err := cred.Get()
//	if err != nil {
//		log.Println("S3Service::initSingleClient, init credentials failed, err:", err.Error())
//		return nil
//	}
//
//	//init aws config
//	cfg := aws.NewConfig().WithRegion(region).WithCredentials(cred)
//	if cfg == nil {
//		log.Println("S3Service::initSingleClient, init config failed")
//		return nil
//	}
//
//	//init s3 client
//	svc := s3.New(session.New(), cfg)
//
//	return svc
//}
//
//
////////////////////////////
////private func for worker
////////////////////////////
//
////sub async worker process
//func (w *S3Worker) runMainProcess() {
//	var (
//		s3Data S3Data
//		needQuit bool
//		respBool bool
//		respByte = make([]byte, 0)
//	)
//	for {
//		if needQuit && len(w.dataChan) <= 0 {
//			break
//		}
//		select {
//		case s3Data = <- w.dataChan:
//			//do diff opt by data kind
//			switch s3Data.kind {
//			case S3DataKindRead:
//				respByte = w.readFile(s3Data.filePath)
//				s3Data.receiverChan <- respByte
//			case S3DataKindSave:
//				respBool = w.saveFile(s3Data.filePath, s3Data.dataBuff)
//				s3Data.notifyChan <- respBool
//			}
//		case <- w.closeChan:
//			needQuit = true
//		}
//	}
//}
//
//
////copy file to new avatar dir
//func (w *S3Worker) copyFile(fileSource, fileTarget string) bool {
//	if fileSource == "" || fileTarget == "" {
//		return false
//	}
//
//	//init copy object
//	copyParam := &s3.CopyObjectInput{
//		Bucket:aws.String(w.bucket),
//		CopySource: aws.String(fileSource),
//		Key:        aws.String(fileTarget),
//	}
//	_, err := w.client.CopyObject(copyParam)
//	if err != nil {
//		log.Println("S3Worker::copyFile, worker:", w.id, " copy source file ", fileSource, " failed, err:", err.Error())
//		return false
//	}
//	return true
//}
//
//
////read file data from s3 service
//func (w *S3Worker) readFile(filePath string) []byte {
//	var (
//		readSize int
//		//fileSize int64
//		fileData = make([]byte, 0)
//		err error
//	)
//
//	//format file path
//	//log.Println("S3Worker::readFile, filePath:", filePath)
//	if filePath == "" {
//		return fileData
//	}
//
//	//init get parameter
//	getParam := &s3.GetObjectInput{
//		Bucket: aws.String(w.bucket),
//		Key: aws.String(filePath),
//	}
//
//	//begin get data from s3 service
//	getResp, err := w.client.GetObject(getParam)
//	if err != nil {
//		log.Println("S3Worker::readFile, read file ", filePath, " failed, err:", err.Error())
//		return []byte{}
//	}
//
//	//get file info
//	//fileSize = *getResp.ContentLength
//	tempBuff := make([]byte, fileTempBuffSie)
//	radSize := 0
//	for {
//		readSize, err = getResp.Body.Read(tempBuff)
//		if readSize <= 0 {
//			break
//		}
//		fileData = append(fileData, tempBuff[0:readSize]...)
//		radSize += readSize
//	}
//	if len(fileData) <= 0 {
//		tempBuff = tempBuff[:0]
//		return []byte{}
//	}
//	return fileData
//}
//
//
////save file data into s3 service
//func (w *S3Worker) saveFile(filePath string,  fileData[]byte) bool {
//	var fileSize int64
//	if filePath == "" || len(fileData) <= 0 {
//		return false
//	}
//
//	//analyze file data
//	fileBytes := bytes.NewReader(fileData)
//	fileType := http.DetectContentType(fileData)
//	fileSize = int64(len(fileData))
//
//	params := &s3.PutObjectInput{
//		Bucket: aws.String(w.bucket),
//		Key: aws.String(filePath),
//		Body: fileBytes,
//		ContentLength: aws.Int64(fileSize),
//		ContentType: aws.String(fileType),
//	}
//
//	//save data to s3 service
//	_, err := w.client.PutObject(params)
//	if err != nil {
//		log.Println("S3Worker::saveFile, worker:", w.id, " save file ", filePath, " failed, err:", err.Error())
//		return false
//	}
//	return true
//}
