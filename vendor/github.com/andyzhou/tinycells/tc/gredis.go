package tc

import (
	"github.com/go-redis/redis"
	"sync"
)

/*
 * Redis client another
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * base on `github.com/go-redis/redis`
 */

 //GRedis info
 type GRedis struct {
	 address string `redis server address`
	 password string `redis server password`
	 db int `db number`
	 poolSize int `redis pool size`
	 client *redis.Client
	 pubSub *redis.PubSub
	 sync.RWMutex
 }

 //construct
func NewGRedis(address, password string, db int) *GRedis {
	return NewGRedisWithPoolSize(address, password, db, DefaultPoolSize)
}

func NewGRedisWithPoolSize(address, password string, db, poolSize int) *GRedis {
	this := &GRedis{
		address:address,
		password:password,
		db:db,
		poolSize:poolSize,
	}
	//inter init
	this.interInit()
	return this
}

//get client
func (r *GRedis) GetClient() *redis.Client {
	return r.client
}

//get pub/sub object
func (r *GRedis) GetPubSub() *redis.PubSub {
	return r.pubSub
}

//set pub/sub object
func (r *GRedis) SetPubSub(obj *redis.PubSub) bool {
	if obj == nil {
		return false
	}
	r.Lock()
	defer r.Unlock()
	r.pubSub = obj

	return true
}

////////////////
//private func
////////////////

//internal init
func (r *GRedis) interInit() {
	redisOption := &redis.Options{
		Addr:r.address,
		Password:r.password,
		DB:r.db,
		PoolSize:r.poolSize,
	}
	r.client = redis.NewClient(redisOption)
}

