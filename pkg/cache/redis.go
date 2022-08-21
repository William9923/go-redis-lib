package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// ErrNil error return no result from cache
var ErrNil = redis.Nil
var ErrEmpty = fmt.Errorf("redis value empty")

const ErrorFailedConnect = "Failed to connect to redis %s. Error: %s"

type Cache interface {
	Ping() error

	Set(ctx context.Context, key string, val string) Reply
	SetWithExpire(ctx context.Context, key string, val string, expire time.Duration) Reply
	SetStruct(ctx context.Context, key string, val interface{}) Reply
	SetStructWithExpire(ctx context.Context, key string, val interface{}, expire time.Duration) Reply

	Get(ctx context.Context, key string) Reply
	GetStruct(ctx context.Context, key string, res interface{}) error

	Exists(ctx context.Context, key string) (bool, error)
	Do(ctx context.Context, command string, args ...interface{}) Reply
}

type Reply interface {
	Err() error
	Val() interface{}
	Text() (string, error)
	Result() (interface{}, error)
}

type Config struct {
	Address  string
	Password string
	DB       int
}

type Redis struct {
	cache *redis.Client
}

type CustomReply struct {
	err       error
	rawResult string
	val       interface{}
}

func (r *CustomReply) Err() error {
	return r.err
}

func (r *CustomReply) Val() interface{} {
	return r.val
}

func (r *CustomReply) Text() (string, error) {
	return fmt.Sprint(r.val), nil
}

func (r *CustomReply) Result() (interface{}, error) {
	return r.val, r.err
}

func Connect(cfg Config) Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Redis{
		cache: rdb,
	}
}

// Do implements Cache
func (r *Redis) Do(ctx context.Context, command string, args ...interface{}) Reply {
	cmd := r.cache.Do(ctx, command, args)
	return cmd
}

// Ping implements Cache
func (r *Redis) Ping() error {
	res := r.cache.Do(context.Background(), "PING").String()
	if res != "PONG" {
		return fmt.Errorf(ErrorFailedConnect, r.cache, "unable to ping redis server")
	}
	return nil
}

// Exists implements Cache
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	reply, err := r.cache.Do(ctx, "EXISTS", key).Int()
	if err != nil {
		return false, fmt.Errorf(ErrorFailedConnect, r.cache, err)
	}

	return reply == 1, nil
}

// Get implements Cache
func (r *Redis) Get(ctx context.Context, key string) Reply {
	cmd := r.cache.Do(ctx, "GET", key)
	return cmd
}

// GetStruct implements Cache
func (r *Redis) GetStruct(ctx context.Context, key string, res interface{}) error {
	val, err := r.Get(ctx, key).Text()
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(val), res); err != nil {
		return err
	}
	return nil
}

// Set implements Cache
func (r *Redis) Set(ctx context.Context, key string, val string) Reply {
	reply := r.cache.Do(ctx, "SET", key, val)
	r.cache.Expire(ctx, key, 15*time.Hour)
	return reply
}

// SetWithExpire implements Cache
func (r *Redis) SetWithExpire(ctx context.Context, key string, val string, expire time.Duration) Reply {
	reply := r.cache.Do(ctx, "SET", key, val)
	r.cache.Expire(ctx, key, expire)
	return reply
}

// SetStruct implements Cache
func (r *Redis) SetStruct(ctx context.Context, key string, val interface{}) Reply {
	res, reply := r.serializer(val)
	if reply != nil {
		return reply
	}
	return r.Set(ctx, key, res)
}

// SetStructWithExpire implements Cache
func (r *Redis) SetStructWithExpire(ctx context.Context, key string, val interface{}, expire time.Duration) Reply {

	res, reply := r.serializer(val)
	if reply != nil {
		return reply
	}
	return r.SetWithExpire(ctx, key, res, expire)
}

// GetSessionUserID implements Cache
func (r *Redis) GetSessionUserID(ctx context.Context, sessionKey string) (int64, error) {
	res, err := r.cache.Do(ctx, "GET", sessionKey).Int64()
	if err != nil {
		return -1, err
	}
	return res, nil
}

// TTL implements Cache
func (r *Redis) TTL(ctx context.Context, key string) Reply {
	return r.cache.Do(ctx, "TTL", key)
}

func (r *Redis) serializer(val interface{}) (string, Reply) {

	valSerialized, err := json.Marshal(val)
	if err != nil {
		return "", &CustomReply{
			err: err,
		}
	}
	return string(valSerialized), nil
}
