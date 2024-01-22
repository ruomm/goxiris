/**
 * @copyright 像衍科技-idr.ai
 * @author 牛牛-研发部-www.ruomm.com
 * @create 2024/1/16 10:22
 * @version 1.0
 */
package xredis

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/ruomm/goxframework/gox/refx"
	"strconv"
	"time"
)

type XRedisClient struct {
	// 环境变量信息，用于生成特定的环境值
	AppEnv string `xref:"AppEnv"`
	//# 主机名称
	Host string `xref:"Host"`
	//# 端口
	Port int `xref:"Port"`
	//# 密码
	Password string `xref:"Password"`
	// # 自定义客户端名称
	ClientName string `xref:"ClientName"`
	//# 使用的数据库
	DbIndex int `xref:"DbIndex"`
	//# 最大闲置连接数
	MaxIdle int `xref:"MaxIdle"`
	//# 最大活动连接数
	MaxActive int `xref:"MaxActive"`
	//# 闲置过期时间
	IdleTimeout int `xref:"IdleTimeout"`
	//# 连接最长生存时间 如果超过时间会被从链表中删除
	MaxConnLifeTime int `xref:"MaxConnLifeTime"`
	//# If Wait is true and the pool is at the MaxActive limit, then Get() waits
	//# for a connection to be returned to the pool before returning.
	WaitEnable bool `xref:"WaitEnable"`
	RedisPool  *redis.Pool
}

var xRedisClient *XRedisClient

func RedisClientInit(appEnv string, redisConfig interface{}) {
	xClient := XRedisClient{}
	refx.XRefStructCopy(redisConfig, &xClient)
	if appEnv == "" {
		xClient.AppEnv = "dev"
	} else {
		xClient.AppEnv = appEnv
	}
	if xClient.MaxIdle <= 0 || xClient.MaxIdle > 100 {
		xClient.MaxIdle = 5
	}
	if xClient.MaxActive <= 0 || xClient.MaxActive > 100 {
		xClient.MaxActive = 5
	}
	if xClient.IdleTimeout < 0 || xClient.IdleTimeout > 3600*24 {
		xClient.IdleTimeout = 60 * 5
	}
	if xClient.MaxConnLifeTime <= 0 || xClient.MaxConnLifeTime > 3600*24*2 {
		xClient.MaxConnLifeTime = 60 * 10
	}
	xRedisClient = &xClient
}
func SingleRedisClient() *XRedisClient {
	return xRedisClient
}
func (t *XRedisClient) InitPull() {
	var dial func() (redis.Conn, error) = nil
	if t.Password == "" && t.ClientName == "" {
		dial = func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", t.Host+":"+strconv.Itoa(t.Port), redis.DialDatabase(t.DbIndex))
			return conn, err
		}
	} else if t.Password == "" {
		dial = func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", t.Host+":"+strconv.Itoa(t.Port), redis.DialClientName(t.ClientName), redis.DialDatabase(t.DbIndex))
			return conn, err
		}
	} else if t.ClientName == "" {
		dial = func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", t.Host+":"+strconv.Itoa(t.Port), redis.DialPassword(t.Password), redis.DialDatabase(t.DbIndex))
			return conn, err
		}
	} else {
		dial = func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", t.Host+":"+strconv.Itoa(t.Port), redis.DialClientName(t.ClientName), redis.DialPassword(t.Password), redis.DialDatabase(t.DbIndex))
			return conn, err
		}
	}
	t.RedisPool = &redis.Pool{
		MaxIdle:         t.MaxIdle,
		MaxActive:       t.MaxActive,
		Dial:            dial,
		IdleTimeout:     time.Duration(t.IdleTimeout) * time.Second,
		MaxConnLifetime: time.Duration(t.MaxConnLifeTime) * time.Second,
		Wait:            t.WaitEnable,
	}

}
func (t *XRedisClient) Set(key string, value interface{}) error {
	// 执行Redis命令
	conn := t.RedisPool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	_, err := conn.Do("Set", key, value)
	return err
}
func (t *XRedisClient) SetEx(key string, expiresAt int, value interface{}) error {
	// 执行Redis命令
	conn := t.RedisPool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	_, err := conn.Do("setex", key, expiresAt, value)
	return err
}

func (t *XRedisClient) SetObject(key string, value interface{}) error {
	dataJson, errJson := json.Marshal(value)
	if errJson != nil {
		return errJson
	}
	// 执行Redis命令
	conn := t.RedisPool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	_, err := conn.Do("Set", key, string(dataJson))
	return err
}
func (t *XRedisClient) SetExObject(key string, expiresAt int, value interface{}) error {
	dataJson, errJson := json.Marshal(value)
	if errJson != nil {
		return errJson
	}
	// 执行Redis命令
	conn := t.RedisPool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	_, err := conn.Do("setex", key, expiresAt, string(dataJson))
	return err
}

func (t *XRedisClient) GetString(key string) (string, error) {
	// 执行Redis命令
	conn := t.RedisPool.Get()
	defer conn.Close()
	if conn.Err() != nil {
		return "", conn.Err()
	}
	reply, err := redis.String(conn.Do("GET", key))
	return reply, err
}

func (t *XRedisClient) GetObject(key string, v interface{}) error {
	// 执行Redis命令
	conn := t.RedisPool.Get()
	defer conn.Close()
	if conn.Err() != nil {
		return conn.Err()
	}
	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(reply), v)
	return err
}
func (t *XRedisClient) GetInt(key string) (int, error) {
	// 执行Redis命令
	conn := t.RedisPool.Get()
	defer conn.Close()
	if conn.Err() != nil {
		return 0, conn.Err()
	}
	reply, err := redis.Int(conn.Do("GET", key))
	return reply, err
}
func (t *XRedisClient) GetInt64(key string) (int64, error) {
	// 执行Redis命令
	conn := t.RedisPool.Get()
	defer conn.Close()
	if conn.Err() != nil {
		return 0, conn.Err()
	}
	reply, err := redis.Int64(conn.Do("GET", key))
	return reply, err
}
