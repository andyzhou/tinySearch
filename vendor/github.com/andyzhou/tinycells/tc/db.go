package tc

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"sync"
	"time"
)

/*
 * Mysql db service interface, this remove to sub dir of `db`
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //internal macro variables
 const (
 	LazyCommandChanSize = 32
 	ConnCheckRate = 20 //xxx seconds
 	DBPoolMin = 1
 	DBPoolMax = 64
 )

 //internal db server status
 const (
 	DBStateActive = 1
 	DBStateDown = 0
 )

 //lazy command
type LazyCmd struct {
	query string `sql string`
	args []interface{} `dynamic args slice`
}

 //db service info
type DBService struct {
	address string `mysql auth address`
	poolSize int
	dbPool map[int]*sql.DB `db pool`
	db *sql.DB `db connect instance`
	state int `db state`
	lazyChan chan LazyCmd `lazy command chan`
	closeChan chan bool
	err error `error interface`
	sync.Mutex
	Utils
}

//construct
func NewDBService(dbAddress string) *DBService {
	return NewDBServiceWithPool(dbAddress, DBPoolMin)
}

//with pool
func NewDBServiceWithPool(dbAddress string, poolSize int) *DBService {
	if poolSize < DBPoolMin {
		poolSize = DBPoolMin
	}
	if poolSize > DBPoolMax {
		poolSize = DBPoolMax
	}
	address := fmt.Sprintf("%s?charset=utf8", dbAddress)
	this := &DBService{
		address:address,
		poolSize:poolSize,
		dbPool:make(map[int]*sql.DB),
		db:nil,
		state:DBStateDown,
		lazyChan:make(chan LazyCmd, LazyCommandChanSize),
		closeChan:make(chan bool),
	}

	//init db pool
	go this.initPool()

	//start lazy process
	go this.lazyProcess()

	return this
}

//////
//API
//////

//service quit
func (s *DBService) Quit() {
	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("DBService::Quit panic happened, err:", err)
		}
	}()

	s.closeChan <- true
}

//get db instance
func (s *DBService) GetDB() *sql.DB {
	return s.getRandomDB()
}

//execute sql
//return lastInsertId, effectRows, error
func (s *DBService) Execute(query string, args ...interface{}) (int64, int64, error){
	if s.state == DBStateDown {
		return 0, 0, errors.New("DB connect down")
	}
	//get random db
	db := s.getRandomDB()
	if db == nil {
		return 0, 0, errors.New("DB connect is null")
	}
	//exec sql
	result, err := db.Exec(query, args...)
	if err != nil {
		log.Println("Execute sql ", query, " failed, error:", err.Error())
		return 0, 0, err
	}
	lastInsertId, _ := result.LastInsertId()
	effectRows, _ := result.RowsAffected()
	return lastInsertId, effectRows, nil
}

//get one row record
func (s *DBService) GetRow(query string, args ...interface{}) (map[string]interface{}, error) {
	if s.state == DBStateDown {
		return nil, errors.New("DB connect down")
	}
	recordMap := make(map[string]interface{})
	queryNew := fmt.Sprintf("%s LIMIT 1", query)
	records, err := s.GetArray(queryNew, args...)
	if err != nil {
		return nil, err
	}
	//return first record of slice
	for _, record := range records {
		if len(record) <= 0 {
			continue
		}
		recordMap = record
		break
	}
	return recordMap, nil
}

//get batch records
func (s *DBService) GetArray(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if s.state == DBStateDown {
		return nil, errors.New("DB connect down")
	}

	//get random db
	db := s.getRandomDB()
	if db == nil {
		return nil, errors.New("DB connect is null")
	}

	recordsSlice := make([]map[string]interface{}, 0)
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Query %s failed, error:%s\n", query, err.Error())
		return nil, err
	}

	//init map for return
	columns, _ := rows.Columns()

	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	//begin get record and copy into slice for return
	k := 0
	for rows.Next() {
		//storage row data into record map
		err = rows.Scan(scanArgs...)
		record := make(map[string]interface{})
		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		//append record into slice list
		recordsSlice = append(recordsSlice, record)
		k++
	}
	return recordsSlice, nil
}

//lazy execute sql in async way
func (s *DBService) LazyExecute(sql string, args ...interface{}) bool {
	if sql == "" {
		return false
	}
	lazy := LazyCmd{
		query:sql,
		args:make([]interface{}, 0),
	}
	lazy.args = append(lazy.args, args...)
	//cast command to chan
	s.lazyChan <- lazy
	return true
}


/////////////////////
//private function
/////////////////////

//execute command only
func (s *DBService) onlyExecuteLazy(lazy LazyCmd) bool {
	//try get random db
	db := s.getRandomDB()
	if db == nil {
		return false
	}

	//exec sql
	_, err := db.Exec(lazy.query, lazy.args...)
	if err != nil {
		log.Println("onlyExecuteLazy exec failed, error:", err.Error())
		log.Println("args:", lazy.args, ", len:", len(lazy.args))
		//put into queue?
		return false
	}
	return true
}

//lazy process
func (s *DBService) lazyProcess()  {
	var (
		lazy LazyCmd
		needQuit, isOk bool
		tick = time.Tick(time.Second * ConnCheckRate)
	)

	for {
		if needQuit && len(s.lazyChan) <= 0 {
			break
		}
		select {
			case lazy, isOk = <- s.lazyChan:
				if isOk {
					s.onlyExecuteLazy(lazy)
				}
			case <- tick:
				s.checkConnect()
			case <- s.closeChan:
				needQuit = true
		}
	}
}

//get rand db connect
func (s *DBService) getRandomDB() *sql.DB {
	realPoolSize := len(s.dbPool)
	if realPoolSize <= 0 {
		return nil
	}
	randIdx := s.GetRandomVal(realPoolSize) + 1
	//log.Println("DBService::getRandomDB, randIdx:", randIdx, ", realPoolSize:", realPoolSize)
	v, ok := s.dbPool[randIdx]
	if !ok {
		return nil
	}
	return v
}

//check server connect
func (s *DBService) checkConnect() bool {
	var (
		err error
	)
	if s.poolSize <= 0 {
		return false
	}

	//check one by one
	failed := 0
	for idx, db := range s.dbPool {
		err = db.Ping()
		if err == nil {
			//check pass
			continue
		}

		//some error happened
		log.Println("ping db server, err:", err.Error())
		db.Close()

		//try connect db
		bRet, dbNew := s.connectServer()
		if !bRet {
			failed++
			continue
		}
		s.dbPool[idx] = dbNew
	}

	if failed == s.poolSize {
		//all failed, mark down!
		s.state = DBStateDown
	}

	return true
}

//connect db server
func (s *DBService) connectServer() (bool, *sql.DB) {
	var tip string

	//init db driver
	db, err := sql.Open("mysql", s.address)
	if err != nil {
		//connect db failed
		tip = "Connect db server failed, error:" + err.Error()
		log.Println(tip)
		panic(tip)
		return false, nil
	}

	//try ping db server
	err = db.Ping()
	if err != nil {
		tip = "ping db server failed, error:" + err.Error()
		log.Println(tip)
		panic(tip)
		return false, nil
	}

	return true, db
}

//init db pool
func (s *DBService) initPool() {
	var k = 1
	for i := 1; i <= s.poolSize; i++ {
		bRet, db := s.connectServer()
		if !bRet {
			continue
		}
		//add into pool
		s.Lock()
		s.dbPool[k] = db
		s.Unlock()
		k++
	}

	if len(s.dbPool) > 0 {
		s.state = DBStateActive
	}
}
