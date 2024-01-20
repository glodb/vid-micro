package cache

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/sync/semaphore"
)

type RedisCache struct {
	pool      *redis.Pool
	semaphore *semaphore.Weighted
}

var instance *RedisCache
var once sync.Once

func GetInstance() *RedisCache {
	var err error
	once.Do(func() {
		instance = &RedisCache{nil, nil}
		instance.pool, err = instance.newPool()
		instance.semaphore = semaphore.NewWeighted(int64(configmanager.GetInstance().Redis.RedisMaxConnections))
		if err != nil {

			return
		}
	})
	return instance
}

// refresh Pool is used to refresh the redis pool
func (cache *RedisCache) refreshPool() error {
	var err error
	cache.pool, err = cache.newPool()
	return err
}

func (cache *RedisCache) GetConnection() redis.Conn {
	cache.semaphore.Acquire(context.TODO(), 1)
	c := cache.pool.Get()
	return c
}

func (cache *RedisCache) ReleaseConnection(conn redis.Conn) {
	conn.Close()
	cache.semaphore.Release(1)
}

func (cache *RedisCache) newPool() (*redis.Pool, error) {
	var redErr error
	pool := redis.Pool{
		MaxIdle:   configmanager.GetInstance().Redis.RedisMaxIdleConnections,
		MaxActive: configmanager.GetInstance().Redis.RedisMaxConnections, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(configmanager.GetInstance().Redis.RedisCon, configmanager.GetInstance().Redis.RedisAddress)
			if err != nil {
				redErr = err

			}
			return c, err
		},
	}
	return &pool, redErr
}

func (cache *RedisCache) Set(key string, value []byte) error {
	cache.semaphore.Acquire(context.TODO(), 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("SET", key, value)
	if err != nil {

		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {Set} key:", key, " error:", err)
		}
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
	}
	return err
}

func (cache *RedisCache) GetInt(key string) (int64, error) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("TError: ("+configmanager.GetInstance().ServiceLogName+"),acquire semaphore:", err)
		}
		return -1, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	var data int64
	dataint, err := c.Do("GET", key)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {GetInt} key:", key, " error:", err)
		}
		return -1, err
	}
	if dataint != nil {
		data, err = redis.Int64(dataint, err)
	}
	return data, err
}

func (cache *RedisCache) SetInt(key string, value int) error {
	cache.semaphore.Acquire(context.TODO(), 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)

	defer c.Close()
	_, err := c.Do("SET", key, value)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {SetInt} key:", key, " val:", value, " error:", err)
		}
	}
	return err
}

func (cache *RedisCache) Increment(key string) (int64, error) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return -1, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)

	defer c.Close()
	val, err := c.Do("INCR", key)
	if err != nil {

		return -1, err
	}
	return val.(int64), err
}

func (cache *RedisCache) Decrement(key string) (int64, error) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return -1, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)

	defer c.Close()
	val, err := c.Do("DECR", key)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {Decrement} key:", key, " err:", err)
		}
		return -1, err
	}
	return val.(int64), err
}

func (cache *RedisCache) SetString(key string, value string) error {
	cache.semaphore.Acquire(context.TODO(), 1)

	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("SET", key, value)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {SetString} key:", key, " val:", value, " err:", err)
		}
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
	}
	return err
}

func (cache *RedisCache) Get(key string) ([]byte, error) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return nil, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	var data []byte
	dataint, err := c.Do("GET", key)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {Get} key:", key, " err:", err)
		}
		return []byte{}, err
	}
	if dataint != nil {
		data, err = redis.Bytes(dataint, err)
	}
	return data, err
}

func (cache *RedisCache) GetKeys(pattern string) []string {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	data, _ := redis.Strings(c.Do("Keys", pattern))
	return data
}
func (cache *RedisCache) GetString(key string) (string, error) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return "", err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	var data string
	dataint, err := c.Do("GET", key)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {GetString} key:", key, " err:", err)
		}
		return "", err
	}
	if dataint != nil {
		data, err = redis.String(dataint, err)
	}
	return data, err
}

func (cache *RedisCache) Del(key string) error {
	cache.semaphore.Acquire(context.TODO(), 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("DEL", key)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {Del} key:", key, " err:", err)
		}
	}
	return err
}

func (cache *RedisCache) Append(key string, value interface{}) error {
	cache.semaphore.Acquire(context.TODO(), 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("APPEND", key, value)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {Append} key:", key, " val:", value, " err:", err)
		}
	}
	return err
}

func (cache *RedisCache) SAdd(value []interface{}) error {
	if len(value) <= 1 {
		return errors.New("Not enough parameters")
	}
	cache.semaphore.Acquire(context.TODO(), 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("SADD", value...)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {SAdd} key:", value, " err:", err)
		}
	}
	return err
}

func (cache *RedisCache) SMembers(key string) []string {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("SMEMBERS", key))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {ReadSet} key:", key, " err:", err)
		}
	}
	return data
}
func (cache *RedisCache) SRem(value []interface{}) error {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("SREM", value...)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {ReadSet} key:", value, " err:", err)
		}
	}
	return err
}

func (cache *RedisCache) SetISMember(key string, member string) (bool, error) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return false, nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	val, err := redis.Bool(c.Do("SISMEMBER", key, member))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {SetISMember} key:", key, "member:", member, " err:", err)
		}
		return false, nil
	}
	return val, err
}

func (cache *RedisCache) SortedSetAdd(key string, seq int32, value interface{}) error {
	cache.semaphore.Acquire(context.TODO(), 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("ZADD", key, seq, value)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {ZAdd} key:", key, " err:", err)
		}
	}
	return err
}

func (cache *RedisCache) ReadSortedSet(key string) []string {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

		return nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {ReadSortedSet} key:", key, " err:", err)
		}
	}
	return data
}

func (cache *RedisCache) RemoveSortedMessage(key string, val string) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("ZREM", key, val)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {RemoveSortedMessage} key:", key, " error:", err)
		}
	}
}

func (cache *RedisCache) HashMultiSet(key string, args map[string]interface{}) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HMSET", redis.Args{key}.AddFlat(args)...)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashMultiSet} key:", key, " error:", err)
		}
	}
}

func (cache *RedisCache) HashMultiSetString(key string, args map[string]string) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HMSET", redis.Args{key}.AddFlat(args)...)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashMultiSet} key:", key, " error:", err)
		}
	}
}

func (cache *RedisCache) HashMultiSetInt(key string, args map[string]int) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HMSET", redis.Args{key}.AddFlat(args)...)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashMultiSetInt} key:", key, " args", args, " error:", err)
		}
	}
}

// HashSet first index should be key
func (cache *RedisCache) HashSet(args []interface{}) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HSET", args...)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashSet} key:", args, " err:", err)
		}
	}
}

func (cache *RedisCache) LPush(key string, args []byte) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {
		return
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("LPUSH", key, args)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {LPUSH} key:", args, " err:", err)
		}
	}
}

func (cache *RedisCache) LTrim(key string, start int, end int) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {
		return
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("LTRIM", key, start, end)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {LTRIM} key:", key, " err:", err)
		}
	}
}

func (cache *RedisCache) LRange(key string, start int, end int) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {
		return
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := redis.Strings(c.Do("LRANGE", key, start, end))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {LTRIM} key:", key, " err:", err)
		}
	}
}

func (cache *RedisCache) HashGet(key string, field string) string {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.String(c.Do("HGET", key, field))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashGet} key:", key, ",Field:", field, "err:", err)
		}
	}
	return data
}

func (cache *RedisCache) HashGetNoPrint(key string, field string) (string, error) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.String(c.Do("HGET", key, field))
	return data, err
}

func (cache *RedisCache) HashGetBytes(key string, field string) []byte {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Bytes(c.Do("HGET", key, field))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashGet} key:", key, ",Field:", field, "err:", err)
		}
	}
	return data
}
func (cache *RedisCache) HashDel(key string, field string) error {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HDEL", key, field)
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashDel} key:", key, " err:", err)
		}
	}
	return err
}

func (cache *RedisCache) HashGetAll(key string) map[string]string {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.StringMap(c.Do("HGETALL", key))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashGetAll} key:", key, " err:", err)
		}
	}
	return data
}

func (cache *RedisCache) HashGetInt(key string, field string) int {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Int(c.Do("HGET", key, field))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashGetInt} key:", key, " error:", err)
		}
	}
	return data
}

func (cache *RedisCache) HashGetBool(key string, field string) bool {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Bool(c.Do("HGET", key, field))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashGetBool} key:", key, " error:", err)
		}
	}
	return data
}

func (cache *RedisCache) HashGetInt64(key string, field string) int64 {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Int64(c.Do("HGET", key, field))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashGetInt64} key:", key, " err:", err)
		}
	}
	return data
}

func (cache *RedisCache) HashGetFloat64(key string, field string) float64 {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Float64(c.Do("HGET", key, field))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashGetInt64} key:", key, " err:", err)
		}
	}
	return data
}

func (cache *RedisCache) HashGetInt64WithError(key string, field string) (int64, error) {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Int64(c.Do("HGET", key, field))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HashGetInt64} key:", key, " err:", err)
		}
	}
	return data, err
}

func (cache *RedisCache) HashMGet(field []interface{}) []string {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("HMGET", field...))
	if err != nil {
		if configmanager.GetInstance().Redis.PrintRedis {
			log.Println("RError: ("+configmanager.GetInstance().ServiceLogName+"), {HMGET} key:", field, " err:", err)
		}
	}
	return data
}

func (cache *RedisCache) HashMGetInts(field []interface{}) []int {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Ints(c.Do("HMGET", field...))
	if err != nil {

	}
	return data
}

func (cache *RedisCache) HashIncrementBy(key string, field string, val int) int64 {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	value, err := redis.Int64(c.Do("HINCRBY", key, field, val))
	if err != nil {

	}
	return value
}
func (cache *RedisCache) HashIncrementByFloat(key string, field string, val float64) int64 {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	value, err := redis.Int64(c.Do("HINCRBYFLOAT", key, field, val))
	if err != nil {

	}
	return value
}

func (cache *RedisCache) Exists(key string) bool {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Bool(c.Do("EXISTS", key))
	if err != nil {

	}
	return data
}

func (cache *RedisCache) GetMinSortedSet(key string) int {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES", "LIMIT", 0, 1))
	if err != nil {

	}
	if len(data) != 0 {
		val, _ := strconv.Atoi(data[1])
		return val
	}
	return 0
}

func (cache *RedisCache) GetMaxSortedSet(key string) int {
	if err := cache.semaphore.Acquire(context.TODO(), 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZRANGEBYSCORE", key, "+inf", "-inf", "WITHSCORES", "LIMIT", 0, 1))
	if err != nil {

	}
	if len(data) != 0 {
		val, _ := strconv.Atoi(data[1])
		return val
	}
	return 0
}
