package app

import (
	"database/sql"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
	"math/rand"
	"strconv"
	"strings"
	"sync"
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
	SELECT_PUSHCNT          = "select id,customer_id,tenant_id,caller_id,type,push_type,count,sharding,created,updated from push_cnt where customer_id=? and tenant_id=? and caller_id = ? and type=? and push_type=? and sharding =?"
	INSERT_PUSHCNT          = "insert into push_cnt( id,customer_id,tenant_id,caller_id,type,push_type,count,sharding,created,updated) values(?,?,?,?,?,?,?,?,?,?)"
	ADD_PUSH_COUNT          = "update push_cnt set count = count+?  where  customer_id=? and tenant_id=? and caller_id = ? and type=? and push_type=? and sharding =?"
	GET_PUSH_COUNT          = "select sum(count) as count from push_cnt WHERE  customer_id  = ?  tenantId = ?"
	GET_CUSTOMER_PUSH_COUNT = "select sum(count) as count from push_cnt  WHERE  customer_id  = ?  "
	PUSH_CNT                = "push_cnt"
)

var lock sync.Mutex

//统计消息推送
func StatisticsPush(pushCnt *PushCnt) {

	if len(pushCnt.Type) > 0 {
		pushCnt.Type = strings.ToLower(pushCnt.Type)
	}

	if DEVICE_TYPE_ANDROID == pushCnt.Type || DEVICE_TYPE_IOS == pushCnt.Type {
		pushCnt.Sharding = 1
	} else {
		pushCnt.Sharding = rand.Intn(10)
		pushCnt.Type = "app"
	}
	/*
		//获取租户信息
		tenant := getTenantById(pushCnt.TenantId)
		if tenant == nil {
			glog.Error("not found tenant")
			return
		}
		pushCnt.CustomerId = tenant.CustomerId
	*/
	//同步问题
	lock.Lock()
	if isExistPushCnt(pushCnt) { //update
		addPushCount(pushCnt)
	} else { //insert
		insertPushCnt(pushCnt)
	}
	lock.Unlock()
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

	_, err = tx.Exec(ADD_PUSH_COUNT, pushCnt.Count, pushCnt.CustomerId, pushCnt.TenantId, pushCnt.CallerId, pushCnt.Type, pushCnt.PushType, pushCnt.Sharding)
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

func getPushCount(customerId, tenantId string) int {
	count := 0
	var rows *sql.Rows
	var err error
	//查询customer下的所有租户调用次数
	if tenantId == "*" {
		rows, err = db.MySQL.Query(GET_CUSTOMER_PUSH_COUNT, customerId)
	} else {
		rows, err = db.MySQL.Query(GET_PUSH_COUNT, customerId, tenantId)
	}

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		glog.Error(err)
		return count
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
		return count
	}
	//取出count
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			//glog.Error(err)
			return count
		}
		return count
	}
	return count
}

func ValidPush(pushCnt *PushCnt) bool {
	//校验时间期限
	key := pushCnt.CustomerId + pushCnt.TenantId + EXPIRE
	quota, ok := QuotaAll[key]
	//该租户不存在配置侧查询全局配置
	if !ok {
		key := pushCnt.CustomerId + "*" + EXPIRE
		quota, ok = QuotaAll[key]
	}
	if ok {
		validTime, err := time.ParseInLocation("2006/01/02 15:04:05", quota.Value, time.Local)
		if err != nil || validTime.Before(time.Now().Local()) {
			glog.Error(err)
			return false
		}
		//校验api调用计数
		key = pushCnt.CustomerId + pushCnt.TenantId + PUSH_CNT
		quota, ok = QuotaAll[key]
		if !ok {
			key = pushCnt.CustomerId + "*" + PUSH_CNT
			quota, ok = QuotaAll[key]
		}
		if ok {
			quotaCount, err := strconv.Atoi(quota.Value)
			if err != nil {
				glog.Error(err)
				return false
			}
			if quotaCount == -1 { //-1 表示不限限制次数
				return true
			}
			count := getPushCount(quota.CustomerId, quota.TenantId)
			if quotaCount > count {
				return true
			}
		} else {
			return false
		}

	} else {
		return false
	}
	return false
}
