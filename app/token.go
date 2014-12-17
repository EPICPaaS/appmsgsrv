package app

import (
	"encoding/json"
	"errors"
	"github.com/EPICPaaS/appmsgsrv/ketama"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"net/http"
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
	logger.Info("Connecting Redis....")

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
			logger.Errorf("strings.Split(\"%s\", :) failed (%v)", n, err)
			panic(err)
		}

		w, err = strconv.Atoi(nw[1])
		if err != nil {
			logger.Errorf("strconv.Atoi(\"%s\") failed (%v)", nw[1], err)
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
					logger.Errorf("redis.Dial(\"tcp\", \"%s\") error(%v)", tmp, err)
					return nil, err
				}

				return conn, err
			},
		}

		ring.AddNode(nw[0], w)
	}

	ring.Bake()
	rs = &redisStorage{pool: redisPool, ring: ring}

	logger.Info("Redis connected")
}

// 根据令牌返回用户. 如果该令牌可用，刷新其过期时间.
func getUserByToken(token string) *member {
	conn := rs.getConn("token")

	if conn == nil {
		return nil
	}

	defer conn.Close()

	if err := conn.Send("EXISTS", token); err != nil {
		logger.Error(err)
	}

	if err := conn.Flush(); err != nil {
		logger.Error(err)
	}

	reply, err := conn.Receive()
	if err != nil {
		logger.Error(err)

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
		logger.Error(err)
	}

	if err := conn.Flush(); err != nil {
		logger.Error(err)
	}

	_, err = conn.Receive()
	if err != nil {
		logger.Error(err)
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
		logger.Error(err)
		return "", err
	}

	// 设置令牌过期时间
	if err := conn.Send("EXPIRE", token, confExpire); err != nil {
		logger.Error(err)
		return "", err
	}

	if err := conn.Flush(); err != nil {
		logger.Error(err)
		return "", err
	}

	_, err := conn.Receive()
	if err != nil {
		logger.Error(err)
		return "", err
	}

	_, err = conn.Receive()
	if err != nil {
		logger.Error(err)
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
		logger.Warnf("key: \"%s\" hit redis node: \"%s\" not in pool", key, node)
		return nil
	}

	logger.Tracef("key: \"%s\" hit redis node: \"%s\"", key, node)

	return p.Get()
}

/*移除 token*/
func (*device) DelToken(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	baseRes := baseResponse{OK, ""}
	body := ""
	res := map[string]interface{}{"baseResponse": &baseRes}
	defer RetPWriteJSON(w, r, res, &body, time.Now())

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = ParamErr
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)

	var args map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &args); err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = ParamErr
		return
	}

	baseReq := args["baseRequest"].(map[string]interface{})
	currentToken, _ := baseReq["token"].(string)
	deltoken, ok := args["delToken"].(string)
	if !ok {
		baseRes.Ret = ParamErr
		return
	}

	//token校验
	user := getUserByToken(currentToken)
	if nil == user {
		baseRes.Ret = AuthErr
		return
	}
	/*删除推token*/
	if delToken(deltoken) {
		return
	} else {
		baseRes.Ret = InternalErr
		baseRes.ErrMsg = "delete token fail "
	}

}

/*移除 token*/
func (*device) ValidateToken(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	baseRes := baseResponse{OK, ""}
	body := ""
	res := map[string]interface{}{"baseResponse": &baseRes}
	defer RetPWriteJSON(w, r, res, &body, time.Now())

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = ParamErr
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)

	var args map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &args); err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = ParamErr
		return
	}

	validToken, ok := args["validToken"].(string)
	if !ok {
		baseRes.Ret = ParamErr
		return
	}
	//token校验
	user := getUserByToken(validToken)
	if nil == user {
		baseRes.Ret = NotFoundServer
		return
	}

	return

}

/*删除token*/
func delToken(token string) bool {

	conn := rs.getConn("token")
	if conn == nil {
		logger.Info(RedisNoConnErr)
		return false
	}
	defer conn.Close()

	// 使用 Redis Hash 结构保存用户令牌值
	if err := conn.Send("DEL", token); err != nil {
		logger.Error(err)
		return false
	}

	if err := conn.Flush(); err != nil {
		logger.Error(err)
		return false
	}

	_, err := conn.Receive()
	if err != nil {
		logger.Error(err)
		return false
	}

	return true
}
