package cache

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/minghsu0107/saga-account/config"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var (
	RedisClient redis.UniversalClient
	//ErrRedisUnlockFail is redis unlock fail error
	ErrRedisUnlockFail = errors.New("redis unlock fail")
	// ErrRedisPipelineCmdNotFound is redis command not found error
	ErrRedisPipelineCmdNotFound = errors.New("redis pipeline command not found; supports only SET and DELETE")
)

// RedisCache is the interface of redis cache
type RedisCache interface {
	Get(ctx context.Context, key string, dst interface{}) (bool, error)
	Set(ctx context.Context, key string, val interface{}) error
	Delete(ctx context.Context, key string) error
	GetMutex(mutexname string) *redsync.Mutex
	ExecPipeLine(ctx context.Context, cmds *[]RedisCmd) error
	Publish(ctx context.Context, topic string, payload interface{}) error
}

// RedisCacheImpl is the redis cache client type
type RedisCacheImpl struct {
	client     redis.UniversalClient
	rs         *redsync.Redsync
	expiration int64
}

// RedisOpType is the redis operation type
type RedisOpType int

const (
	// SET represents set operation
	SET RedisOpType = iota
	// DELETE represents delete operation
	DELETE
)

// RedisPayload is a abstract interface for payload type
type RedisPayload interface {
	Payload()
}

// RedisSetPayload is the payload type of set method
type RedisSetPayload struct {
	RedisPayload
	Key string
	Val interface{}
}

// RedisDeletePayload is the payload type of delete method
type RedisDeletePayload struct {
	RedisPayload
	Key string
}

// Payload implements abstract interface
func (RedisSetPayload) Payload() {}

// Payload implements abstract interface
func (RedisDeletePayload) Payload() {}

// RedisCmd represents an operation and its payload
type RedisCmd struct {
	OpType  RedisOpType
	Payload RedisPayload
}

// RedisPipelineCmd is redis pipeline command type
type RedisPipelineCmd struct {
	OpType RedisOpType
	Cmd    interface{}
}

func NewRedisClient(config *config.Config) (redis.UniversalClient, error) {
	RedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:         getServerAddrs(config.RedisConfig.Addrs),
		Password:      config.RedisConfig.Password,
		PoolSize:      config.RedisConfig.PoolSize,
		MaxRetries:    config.RedisConfig.MaxRetries,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	ctx := context.Background()
	pong, err := RedisClient.Ping(ctx).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	}
	redisotel.InstrumentTracing(RedisClient)
	config.Logger.ContextLogger.WithField("type", "setup:redis").Info("successful redis connection: " + pong)
	return RedisClient, nil
}

// NewRedisCache is the factory of redis cache
func NewRedisCache(config *config.Config, client redis.UniversalClient) RedisCache {
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &RedisCacheImpl{
		client:     client,
		rs:         rs,
		expiration: config.RedisConfig.ExpirationSeconds,
	}
}

// Get returns true if the key already exists and set dst to the corresponding value
func (rc *RedisCacheImpl) Get(ctx context.Context, key string, dst interface{}) (bool, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		json.Unmarshal([]byte(val), dst)
	}
	return true, nil
}

// Set sets a key-value pair
func (rc *RedisCacheImpl) Set(ctx context.Context, key string, val interface{}) error {
	strVal, err := json.Marshal(val)
	if err != nil {
		return err
	}
	if err := rc.client.Set(ctx, key, strVal, getRandomExpiration(rc.expiration)).Err(); err != nil {
		return err
	}
	return nil
}

// Delete deletes a key
func (rc *RedisCacheImpl) Delete(ctx context.Context, key string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (rc *RedisCacheImpl) GetMutex(mutexname string) *redsync.Mutex {
	return rc.rs.NewMutex(mutexname, redsync.WithExpiry(5*time.Second))
}

// ExecPipeLine execute the given commands in a pipline
func (rc *RedisCacheImpl) ExecPipeLine(ctx context.Context, cmds *[]RedisCmd) error {
	pipe := rc.client.Pipeline()
	var pipelineCmds []RedisPipelineCmd
	for _, cmd := range *cmds {
		switch cmd.OpType {
		case SET:
			strVal, err := json.Marshal(cmd.Payload.(RedisSetPayload).Val)
			if err != nil {
				return err
			}
			pipelineCmds = append(pipelineCmds, RedisPipelineCmd{
				OpType: SET,
				Cmd:    pipe.Set(ctx, cmd.Payload.(RedisSetPayload).Key, strVal, getRandomExpiration(rc.expiration)),
			})
		case DELETE:
			pipelineCmds = append(pipelineCmds, RedisPipelineCmd{
				OpType: DELETE,
				Cmd:    pipe.Del(ctx, cmd.Payload.(RedisDeletePayload).Key),
			})
		default:
			return ErrRedisPipelineCmdNotFound
		}
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	for _, executedCmd := range pipelineCmds {
		switch executedCmd.OpType {
		case SET:
			if err := executedCmd.Cmd.(*redis.StatusCmd).Err(); err != nil {
				return err
			}
		case DELETE:
			if err := executedCmd.Cmd.(*redis.IntCmd).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rc *RedisCacheImpl) Publish(ctx context.Context, topic string, payload interface{}) error {
	strVal, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return rc.client.Publish(ctx, topic, strVal).Err()
}

func getRandomExpiration(expiration int64) time.Duration {
	return time.Duration(expiration+rand.Int63n(10)) * time.Second
}

func getServerAddrs(addrs string) []string {
	return strings.Split(addrs, ",")
}
