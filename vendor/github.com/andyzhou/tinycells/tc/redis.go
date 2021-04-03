package tc
import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
	"time"
)

/**
 * redis client api
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */


//field of redis config
const (
	RedisConfigFieldTag = "tag"
	RedisConfigFieldHost = "host"
	RedisConfigFieldPort = "port"
)

//internal macro defines
const (
	LazyRedisChanSize = 64
	RedisCheckRate = 30
	DefaultConnTimeOut = 5
	DefaultDB = 0
	DefaultPoolSize = 1
	DefaultIdx = 1
)

//some special error
const (
	UseLostConnctionErr = "use of closed network connection"
)

//lazy command
type LazyRedisCmd struct {
	command string
	args[]interface{}
}

//single redis info
type Redis struct {
	poolSize int
	address string `redis server address`
	password string `redis server password`
	db int `db number`
	conn *redis.Conn `redis connect instance`
	//connPool map[int]*redis.Conn `redis connect pool map`
	lazyChan chan LazyRedisCmd `lazy command chan`
	lazyStarted bool
	closeChan chan bool
	sync.Mutex `internal locker`
	Utils
}

//running redis pool
type RedisPool struct {
	pool map[string] *Redis `redis pool map, tag:*Redis`
	sync.Mutex `data locker`
}

//construct
func NewRedisPool() *RedisPool {
	this := &RedisPool{
		pool:make(map[string]*Redis),
	}
	return this
}

func NewRedis(address, password string, db int) *Redis {
	return NewRedisWithPoolSize(address, password, db, DefaultPoolSize)
}

func NewRedisWithPoolSize(address, password string, db, poolSize int) *Redis {
	this := &Redis{
		poolSize:poolSize,
		address:address,
		password:password,
		db:db,
		conn:nil,
		//connPool:make(map[int]*redis.Conn),
		lazyChan:make(chan LazyRedisCmd, LazyRedisChanSize),
		closeChan:make(chan bool),
	}
	//start lazy process
	this.startLazyProcess()
	return this
}

////////////////////
//API for RedisPool
////////////////////

//quit
func (rp *RedisPool) Quit() bool {
	if len(rp.pool) <= 0 {
		return false
	}
	for _, redis := range rp.pool {
		redis.Quit()
	}
	return true
}

//cast on all client
func (rp *RedisPool) CastAll(command string, args ...interface{}) bool {
	if len(rp.pool) <= 0 {
		return false
	}
	for _, redis := range rp.pool {
		redis.Exec(command, args...)
	}
	return true
}

//add new redis client
func (rp *RedisPool) AddClient(tag, address string) bool {
	return rp.AddClientWithPassword(tag, address, "", DefaultDB)
}

//add new client with password
func (rp *RedisPool) AddClientWithPassword(tag, address, password string, db int) bool {
	if tag == "" || address == "" {
		return false
	}
	if _, ok := rp.pool[tag]; ok {
		//already exists
		return false
	}

	//add new
	redis := NewRedis(address, password, db)

	//sync into map
	rp.Lock()
	rp.pool[tag] = redis
	rp.Unlock()

	return true
}

//check tag is exists
func (rp *RedisPool) TagIsExists(tag string) bool {
	if len(rp.pool) <= 0 {
		return false
	}
	_, isOk := rp.pool[tag]
	return isOk
}

//get pool by tag
func (rp *RedisPool) GetPool(tag string) *Redis {
	if redis, ok := rp.pool[tag]; ok {
		return redis
	}
	return nil
}

//get rand pool
func (rp *RedisPool) GetRandomPool() *Redis {
	var redis *Redis

	if len(rp.pool) <= 0 {
		return nil
	}

	for _, tmpRedis := range rp.pool {
		redis = tmpRedis
		break
	}
	return redis
}

//execute command on single redis
//on pool
func (rp *RedisPool) DoByTag(tag string, command string, args ...interface{}) (interface{}, error) {
	//get redis by tag
	redis := rp.GetPool(tag)
	if redis == nil {
		return nil, errors.New("no such redis instance")
	}
	return (*redis.conn).Do(command, args...)
}


////////////////
//API for Redis
////////////////

//quit on redis
func (r *Redis) Quit() {
	r.closeChan <- true
	time.Sleep(time.Second/10)
}

//set db number
func (r *Redis) SetDB(db int) bool {
	if db < 0 || r.db == db {
		return false
	}
	r.db = db
	return true
}

//get db number
func (r *Redis) GetDB() int {
	return r.db
}

//publish
func (r *Redis) Publish(channel, message string) error {
	var (
		err error
	)

	//basic check
	if channel == "" || message == "" {
		return errors.New("lost channel or message for publish")
	}

	//get random conn
	//idx, conn := r.getRandomConn()
	idx := 0
	if r.conn == nil {
		return errors.New("No connect of " + r.address)
	}

	//log.Println("Redis::Publish, channel:", channel, ", message:", message)
	_, err = (*r.conn).Do("PUBLISH", channel, message)
	(*r.conn).Flush()
	//log.Println("Redis::Publish, resp:", resp)

	//publish message
	//rep, err := (*r.conn).Do("PUBLISH", channel, message)
	if err != nil {
		log.Println("Redis::Publish, publish failed, err:", err)

		//check lost connect error
		r.checkErrForReConnect(idx, err)

		return err
	}

	return nil
}

//tips:subscribe and publish should use different redis instance!!!!
//subscribe
func (r *Redis) SubscribeSimple(channel string) {
	var (
		err error
	)
	if channel == "" || r.conn == nil {
		return
	}

	err = (*r.conn).Send("SUBSCRIBE", channel)
	if err != nil {
		log.Println("Redis::SubscribeSimple for channel:", channel,
			" failed, err:", err.Error())
		return
	}
	(*r.conn).Flush()

	go func() {
		for {
			reply, err := (*r.conn).Receive()
			if err != nil {
				log.Println("Redis::SubscribeSimple, receive failed, err:", err.Error())
				break
			}
			log.Println("Redis::SubscribeSimple, receive type:", reflect.TypeOf(reply), ", val:", reply)
		}

	}()
}

func (r *Redis) Subscribe(channel string) (*redis.PubSubConn, error) {
	var err error

	if channel == "" {
		return nil, errors.New("lost channel for subscribe")
	}
	//get random conn
	//idx, conn := r.getRandomConn()
	idx := 0
	if r.conn == nil {
		return nil, errors.New("No connect of " + r.address)
	}

	//init pub/sub connect
	psc := redis.PubSubConn{
		Conn:*r.conn,
	}

	//begin subscribe channel
	err = psc.Subscribe(channel)
	if err != nil {
		//check lost connect error
		log.Println("Redis::Subscribe failed, err:", err.Error())
		r.checkErrForReConnect(idx, err)
		return nil, err
	}

	////send SUBSCRIBE command
	//(*r.conn).Send("SUBSCRIBE", channel)
	//(*r.conn).Flush()
	//
	////try receive response
	//for {
	//	reply, err := (*r.conn).Receive()
	//	if err != nil {
	//		log.Println("Redis::Subscribe, err:", err.Error())
	//		break
	//	}
	//	// process pushed message
	//	log.Println("Redis::Subscribe, channel:", channel, ", reply:", reply)
	//	break
	//}
	//
	//return nil, err

	return &psc, nil
}

//do on redis
func (r *Redis) Do(command string, args ... interface{}) (interface{}, error) {
	//get random conn
	//idx, conn := r.getRandomConn()
	idx := 0
	if r.conn == nil {
		return nil, errors.New("No connect of " + r.address)
	}

	//log.Println("Redis::Do, command:", command, ", args:", args)

	//exec with locker
	//r.Lock()
	reply, err := (*r.conn).Do(command, args...)
	//r.Unlock()

	//if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
	//	log.Println("time out..........")
	//}

	if err != nil {
		log.Println("Do...err:", err)
		//reconnect check
		r.checkErrForReConnect(idx, err)
	}

	return reply, err
}

//exec on lazy model
func (r *Redis) Exec(command string, args ...interface{}) bool {
	log.Println("Redis::Exec, lazyStarted:", r.lazyStarted, ", command:", command, ", args:", args)

	if !r.lazyStarted {
		return false
	}
	lazyCommand := LazyRedisCmd{
		command:command,
		args:make([]interface{}, 0),
	}
	lazyCommand.args = append(lazyCommand.args, args...)
	//send to chan
	r.lazyChan <- lazyCommand
	//reset slice
	lazyCommand.args = []interface{}{}
	return true
}

//run script
//keysAndArgs like: key1,key2,arg1,arg2,...
func (r *Redis) RunScript(script string, keyCount int, keysAndArgs ...interface{}) (interface{}, error) {
	if r.conn == nil {
		//try connect
		r.Connect()
	}
	s := redis.NewScript(keyCount, script)
	return s.Do(*r.conn, keysAndArgs)
}


//init new redis connect
func (r *Redis) Connect() (bool, redis.Conn) {
	var (
		conn redis.Conn
		err error
	)
	if r.address == "" {
		return false, nil
	}

	//if connect need password
	conn, err = redis.Dial(
		"tcp",
		r.address,
		redis.DialConnectTimeout(time.Second),
		redis.DialPassword(r.password),
		redis.DialDatabase(r.db),
		redis.DialConnectTimeout(time.Second * DefaultConnTimeOut),
		redis.DialReadTimeout(time.Second),
		redis.DialWriteTimeout(time.Second),
	)

	//try connect redis
	//conn, err := redis.DialTimeout("tcp", r.address, 1*time.Second, 1*time.Second, 1*time.Second)
	if err != nil {
		log.Println("connect redis server ", r.address, " failed, err:", err.Error())
		return false, nil
	}
	//log.Println("connect redis ", r.address, " success")
	return true, conn
}

//////////////////////
//private functions
//////////////////////

//check err for reconnect
func (r *Redis) checkErrForReConnect(idx int, err error) bool {
	var isOk bool

	if err == nil {
		return false
	}

	//special error of net
	netOpError, isOk := err.(*net.OpError)
	if isOk && netOpError.Err.Error() == UseLostConnctionErr {
		//lost connect, try reconnect
		log.Println("try reconnect server...address:", r.address)
		//time.AfterFunc(time.Second/10, r.reconnect)
		r.reconnect(idx)
		return true
	}

	//if connect closed
	if err == io.EOF {
		//lost connect, try reconnect
		log.Println("try reconnect server...address:", r.address)
		//time.AfterFunc(time.Second/10, r.reconnect)
		r.reconnect(idx)
		return true
	}

	return false
}

//run lazy process for single redis
func (r *Redis) startLazyProcess() bool {
	if r.lazyStarted {
		return false
	}

	//init redis pool
	r.initPool()

	//run in sub process
	go r.runLazyProcess()
	r.lazyStarted = true
	return true
}

//reconnect server
func (r *Redis) reconnect(idx int) {
	log.Println("Redis::reconnect........")
	bRet, conn := r.Connect()
	if bRet {
		r.Lock()
		if idx <= 0 {
			r.conn = &conn
		}
		//r.connPool[idx] = &conn
		r.Unlock()
	}
}

//run lazy process for single redis
func (r *Redis) runLazyProcess() {
	var (
		ticker = time.Tick(time.Second * RedisCheckRate)
		lazyCmd LazyRedisCmd
		needQuit bool
	)
	for {
		if needQuit && len(r.lazyChan) <= 0 {
			break
		}
		select {
		case lazyCmd = <- r.lazyChan:
			r.doLazyCmd(lazyCmd)
		case <- ticker:
			//if len(r.connPool) <= 0 {
			//	//re init pool
			//	r.initPool()
			//}
		case <- r.closeChan:
			needQuit = true
		}
	}

	//clean up
	//if len(r.connPool) > 0 {
	//	r.Lock()
	//	for idx, conn := range r.connPool {
	//		(*conn).Close()
	//		delete(r.connPool, idx)
	//	}
	//	r.Unlock()
	//}

	if r.conn != nil {
		(*r.conn).Close()
	}

	//log.Println("runLazyProcess of ", r.address, " has quit")
}

//process single lazy command
func (r *Redis) doLazyCmd(lazyCmd LazyRedisCmd) bool {
	cmd := lazyCmd.command
	args := lazyCmd.args

	log.Println("Redis::doLazyCmd, begin exec...")
	log.Println("Redis::doLazyCmd, command:", cmd, ", args:", args)

	if cmd == "" {
		return false
	}

	//get random connect
	//idx, conn := r.getRandomConn()
	idx := 0
	if r.conn == nil {
		//no connect
		return false
	}

	log.Println("Redis::doLazyCmd, start....")

	//try exceed command
	//r.Lock()
	//err := (*r.conn).Send(cmd, args...)
	reply, err := (*r.conn).Do(cmd, args...)
	log.Println("Redis::doLazyCmd, reply:", reply, ", err:", err)
	if err != nil {
		r.checkErrForReConnect(idx, err)
	}
	//r.Unlock()

	return true
}


////get rand redis connect
//func (r *Redis) getRandomConn() (int, *redis.Conn) {
//	var (
//		bRet bool
//	)
//	realPoolSize := len(r.connPool)
//	if realPoolSize <= 0 {
//		return DefaultIdx, nil
//	}
//	randIdx := r.GetRandomVal(realPoolSize) + 1
//	//log.Println("DBService::getRandomDB, randIdx:", randIdx, ", realPoolSize:", realPoolSize)
//	conn, ok := r.connPool[randIdx]
//	if !ok {
//		return DefaultIdx, nil
//	}
//
//	//check conn, if lost, try re-connect
//	if conn == nil {
//		//try reconnect
//		bRet, *conn = r.Connect()
//		if bRet {
//			//sync into pool map
//			r.Lock()
//			r.connPool[randIdx] = conn
//			r.Unlock()
//		}
//	}
//
//	return randIdx, conn
//}
//
//init redis pool
func (r *Redis) initPool() {
	bRet, conn := r.Connect()
	if bRet {
		r.Lock()
		r.conn = &conn
		r.Unlock()
	}

	//var k = 1
	//for i := 1; i <= r.poolSize; i++ {
	//	//try connect redis server
	//	bRet, conn := r.Connect()
	//	if !bRet {
	//		continue
	//	}
	//	//add into pool
	//	r.Lock()
	//	r.connPool[k] = &conn
	//	r.Unlock()
	//	k++
	//}
}
