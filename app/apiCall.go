package app

import (
	"bytes"
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

const (
	APICALL_EXIST     = "select  id from api_call where customer_id = ? and tenant_id = ? and caller_id=? and type =? and api_name = ? and sharding = ? "
	APICALL_ADD       = "UPDATE api_call SET count = count+1  ,updated=?  WHERE  customer_id = ? and  tenant_id = ? and caller_id=? and type =? and api_name = ? and sharding = ?"
	INSERT_APICALL    = "insert into api_call(id , customer_id,tenant_id,caller_id,type,api_name,count,sharding,created,updated) values(?,?,?,?,?,?,?,?,?,?)"
	GET_APICALL_COUNT = "select sum(count) as count from api_call WHERE  customer_id = ? and  tenant_id = ? and caller_id=? and api_name = ?  "
)

type ApiCall struct {
	Id         string
	CustomerId string
	TenantId   string
	CallerId   string
	Type       string
	ApiName    string
	Count      int
	Sharding   int
	Created    time.Time
	Updated    time.Time
}

//记录api调用次数
func ApiCallStatistics(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	//将body数据写回到request
	rr := bytes.NewReader(body)
	bodyRc := ioutil.NopCloser(rr)
	r.Body = bodyRc

	var args map[string]interface{}
	if err := json.Unmarshal(body, &args); err != nil {
		glog.Errorf(" json.Unmarshal failed (%s)", err)
		return
	}

	baseReq := args["baseRequest"].(map[string]interface{})
	token := baseReq["token"].(string)
	appName := r.URL.String()
	var tenantId, cllerId string
	sharding := 0

	deviceType := "app"
	if baseReq["deviceType"] != nil {
		deviceType = baseReq["deviceType"].(string)
	}
	/* Token 校验，分为用户校验和应用校验*/
	if deviceType == DEVICE_TYPE_IOS || deviceType == DEVICE_TYPE_ANDROID { //移动端和网页端

		var user *member
		if args["userName"] != nil { // 登录api接口

			userName := args["userName"].(string)
			user = getUserByCode(userName)

		} else { //发送消息接口

			user = getUserByToken(token)
		}

		if nil == user {
			glog.V(5).Infof("api_call error: [Logon failure]")
			return
		}

		tenantId = user.TenantId
		cllerId = user.Uid
	} else { //应用校验
		application, err := getApplicationByToken(token)
		if nil != err || nil == application {
			glog.V(5).Infof("api_call error: [Logon failure]")
			return
		}
		//应用10个分片
		sharding = rand.Intn(10)
		tenantId = application.TenantId
		cllerId = application.Id
	}
	//获取租户信息
	tenant := getTenantById(tenantId)
	if tenant == nil {
		glog.Error("not found tenant")
		return
	}
	apiCall := &ApiCall{
		CustomerId: tenant.CustomerId,
		TenantId:   tenantId,
		CallerId:   cllerId,
		Type:       deviceType,
		ApiName:    appName,
		Count:      1, //默认值1
		Sharding:   sharding,
	}

	if apiCallExist(apiCall) { //修改
		addApiCount(apiCall)
	} else { //新增
		insertApiCall(apiCall)
	}

}

//判断该记录是否存在
func apiCallExist(apiCall *ApiCall) bool {

	rows, err := db.MySQL.Query(APICALL_EXIST, apiCall.CustomerId, apiCall.TenantId, apiCall.CallerId, apiCall.Type, apiCall.ApiName, apiCall.Sharding)
	if err != nil {
		glog.Error(err)
		return false
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
		return false
	}
	return rows.Next()
}

//统计调用次数
func addApiCount(apiCall *ApiCall) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}

	_, err = tx.Exec(APICALL_ADD, time.Now().Local(), apiCall.CustomerId, apiCall.TenantId, apiCall.CallerId, apiCall.Type, apiCall.ApiName, apiCall.Sharding)
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

//添加apiCall
func insertApiCall(apiCall *ApiCall) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(INSERT_APICALL, uuid.New(), apiCall.CustomerId, apiCall.TenantId, apiCall.CallerId, apiCall.Type, apiCall.ApiName, apiCall.Count, apiCall.Sharding, time.Now().Local(), time.Now().Local())
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

//获取没用户/应用调用api次数
func getApiCallCount(customerId, tenantId, callerId, apiName string) int {

	count := 0
	rows, err := db.MySQL.Query(GET_APICALL_COUNT, customerId, tenantId, callerId, apiName)
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
			glog.Error(err)
			return count
		}
		return count
	}
	return count
}
