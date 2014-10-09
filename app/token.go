package app

import (
	"errors"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/EPICPaaS/gopush-cluster/ketama"
	"github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
	"strconv"
	"strings"
	"time"
)

var RedisNoConnErr = errors.New("can't get a redis conn")

type redisStorage struct {
	pool map[string]*redis.Pool
	ring *ketama.HashRing
}

var rs *redisStorage

// initRedisStorage initialize the redis pool and consistency hash ring.
func InitRedisStorage() {
	glog.Info("Connecting Redis....")

	var (
		err error
		w   int
		nw  []string
	)

	redisPool := map[string]*redis.Pool{}
	ring := ketama.NewRing(Conf.RedisKetamaBase)

	for n, addr := range Conf.RedisSource {
		nw = strings.Split(n, ":")
		if len(nw) != 2 {
			err = errors.New("node config error, it's nodeN:W")
			glog.Errorf("strings.Split(\"%s\", :) failed (%v)", n, err)
			panic(err)
		}

		w, err = strconv.Atoi(nw[1])
		if err != nil {
			glog.Errorf("strconv.Atoi(\"%s\") failed (%v)", nw[1], err)
			panic(err)
		}

		tmp := addr

		redisPool[nw[0]] = &redis.Pool{
			MaxIdle:     Conf.RedisMaxIdle,
			MaxActive:   Conf.RedisMaxActive,
			IdleTimeout: Conf.RedisIdleTimeout,
			Dial: func() (redis.Conn, error) {
				conn, err := redis.Dial("tcp", tmp)
				if err != nil {
					glog.Errorf("redis.Dial(\"tcp\", \"%s\") error(%v)", tmp, err)
					return nil, err
				}

				return conn, err
			},
		}

		ring.AddNode(nw[0], w)
	}

	ring.Bake()
	rs = &redisStorage{pool: redisPool, ring: ring}

	glog.Info("Redis connected")
}

// 根据令牌返回用户. 如果该令牌可用，刷新其过期时间.
func getUserByToken(token string) *member {
	conn := rs.getConn("token")

	if conn == nil {
		return nil
	}

	defer conn.Close()

	if err := conn.Send("EXISTS", token); err != nil {
		glog.Error(err)
	}

	if err := conn.Flush(); err != nil {
		glog.Error(err)
	}

	reply, err := conn.Receive()
	if err != nil {
		glog.Error(err)

		return nil
	}

	if 0 == reply.(int64) { // 令牌不存在
		return nil
	}

	idx := strings.Index(token, "_")
	if -1 == idx {
		return nil
	}

	uid := token[:idx]

	// 从数据库加载用户
	ret := getUserByUid(uid)
	if nil == ret {
		return nil
	}

	confExpire := int64(Conf.TokenExpire)

	// 刷新令牌
	if err := conn.Send("EXPIRE", token, confExpire); err != nil {
		glog.Error(err)
	}

	if err := conn.Flush(); err != nil {
		glog.Error(err)
	}

	_, err = conn.Receive()
	if err != nil {
		glog.Error(err)
	}

	return ret
}

// 令牌生成.
func genToken(user *member) (string, error) {
	conn := rs.getConn("token")

	if conn == nil {
		return "", RedisNoConnErr
	}

	defer conn.Close()

	confExpire := int64(Conf.TokenExpire)
	expire := confExpire + time.Now().Unix()
	token := user.Uid + "_" + uuid.New()

	// 使用 Redis Hash 结构保存用户令牌值
	if err := conn.Send("HSET", token, "expire", expire); err != nil {
		glog.Error(err)
		return "", err
	}

	// 设置令牌过期时间
	if err := conn.Send("EXPIRE", token, confExpire); err != nil {
		glog.Error(err)
		return "", err
	}

	if err := conn.Flush(); err != nil {
		glog.Error(err)
		return "", err
	}

	_, err := conn.Receive()
	if err != nil {
		glog.Error(err)
		return "", err
	}

	_, err = conn.Receive()
	if err != nil {
		glog.Error(err)
		return "", err
	}

	return token, nil
}

// 获取 Redis 连接.
func (s *redisStorage) getConn(key string) redis.Conn {
	if len(s.pool) == 0 {
		return nil
	}

	node := s.ring.Hash(key)
	p, ok := s.pool[node]
	if !ok {
		glog.Warningf("key: \"%s\" hit redis node: \"%s\" not in pool", key, node)
		return nil
	}

	glog.V(10).Infof("key: \"%s\" hit redis node: \"%s\"", key, node)

	return p.Get()
}
