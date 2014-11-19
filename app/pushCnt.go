package app

import (
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
	"math/rand"
	"time"
)

type PushCnt struct {
	Id         string
	CustomerId string
	TenantId   string
	CallerId   string
	Type       string
	PushType   string
	Count      int
	Sharding   int
	Created    time.Time
	Updated    time.Time
}

const (
	SELECT_PUSHCNT = "select id,customer_id,tenant_id,caller_id,type,push_type,count,sharding,created,updated from push_cnt where customer_id=? and tenant_id=? and type=? and push_type and sharding =?"
	INSERT_PUSHCNT = "insert into push_cnt( id,customer_id,tenant_id,caller_id,type,push_type,count,sharding,created,updated) values(?,?,?,?,?,?,?,?,?,?)"
	ADD_PUSH_COUNT = "update push_cnt set count = count+1  where  customer_id=? and tenant_id=? and type=? and push_type and sharding =?"
)

//统计消息推送
func StatisticsPush(pushCnt *PushCnt) {

	if len(pushCnt.Type) > 0 && (pushCnt.Type == DEVICE_TYPE_ANDROID || pushCnt.Type == DEVICE_TYPE_ANDROID) {
		pushCnt.Sharding = 1
	} else {
		pushCnt.Sharding = rand.Intn(10)
		pushCnt.Type = "app"
	}

	//获取租户信息
	tenant := getTenantById(pushCnt.TenantId)
	if tenant == nil {
		glog.Error("not found tenant")
		return
	}
	pushCnt.CustomerId = tenant.CustomerId
	//update
	if isExistPushCnt(pushCnt) {
		addPushCount(pushCnt)
	} else { //insert
		insertPushCnt(pushCnt)
	}
}

func isExistPushCnt(pushCnt *PushCnt) bool {
	rows, err := db.MySQL.Query(SELECT_PUSHCNT, pushCnt.CustomerId, pushCnt.TenantId, pushCnt.CallerId, pushCnt.Type, pushCnt.PushType, pushCnt.Sharding)
	if err != nil {
		glog.Error(err)
		return false
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		glog.Error(err)
		return false
	}

	return rows.Next()
}

//统计调用次数
func addPushCount(pushCnt *PushCnt) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}

	_, err = tx.Exec(ADD_PUSH_COUNT, pushCnt.CustomerId, pushCnt.TenantId, pushCnt.CallerId, pushCnt.Type, pushCnt.PushType, pushCnt.Sharding)
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}
	//提交操作
	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}
	return true
}

//新增
func insertPushCnt(pushCnt *PushCnt) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(INSERT_PUSHCNT, uuid.New(), pushCnt.CustomerId, pushCnt.TenantId, pushCnt.CallerId, pushCnt.Type, pushCnt.PushType, pushCnt.Count, pushCnt.Sharding, time.Now().Local(), time.Now().Local())
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}
	return true
}
