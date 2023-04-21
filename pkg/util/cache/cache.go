package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type Redis interface {
	Connect(addr, password string, db, pool int)
	Ping() error
	Set(key, val string, expiration int) error
	SetNX(key, val string, expiration int) (bool, error)
	Get(key string) (string, error)
	TTL(key string) (int64, error)
	Exists(key string) (int64, error)
	Del(key string) error
	AcquireLock(key, val string, expiration int) bool
	ReleaseLock(key string) bool
}

type RedisService struct {
	client *redis.Client
}

func NewRedisService() *RedisService {
	return &RedisService{}
}

func (r *RedisService) Connect(addr, password string, db, pool int) {
	r.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: pool,
	})
}

func (r *RedisService) Ping() error {
	return r.client.Ping(context.Background()).Err()
}

func (r *RedisService) Close() error {
	return r.client.Close()
}

func (r *RedisService) Set(key, val string, expiration int) error {
	return r.client.Set(context.Background(), key, val, time.Duration(expiration)*time.Second).Err()
}

func (r *RedisService) SetNX(key, val string, expiration int) (bool, error) {
	result, err := r.client.SetNX(context.Background(), key, val, time.Duration(expiration)*time.Second).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *RedisService) Get(key string) (string, error) {
	val, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisService) TTL(key string) (int64, error) {
	val, err := r.client.TTL(context.Background(), key).Result()
	if err != nil {
		return -1, err
	}
	return int64(val), nil
}

func (r *RedisService) Exists(key string) (int64, error) {
	val, err := r.client.Exists(context.Background(), key).Result()
	if err != nil {
		return -1, err
	}
	return val, nil
}

func (r *RedisService) Del(key string) error {
	return r.client.Del(context.Background(), key).Err()
}

func (r *RedisService) AcquireLock(key, val string, expiration int) bool {
	result, err := r.SetNX(key, val, expiration)
	if err != nil {
		log.Errorf("accuire lock setnx for key: %s failed: %+v", key, err)
		return false
	}
	return result
}

func (r *RedisService) ReleaseLock(key string) bool {
	if err := r.Del(key); err != nil {
		log.Errorf("release lock for key: %s failed: %+v", key, err)
		return false
	}
	return true
}
