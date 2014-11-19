package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	APICALL_EXIST      = "select  id from api_call where customer_id = ? and tenant_id = ? and caller_id=? and type =? and api_name = ? and sharding = ? "
	APICALL_ADD        = "UPDATE api_call SET count = count+1  ,updated=?  WHERE  customer_id = ? and  tenant_id = ? and caller_id=? and type =? and api_name = ? and sharding = ?"
	INSERT_APICALL     = "insert into api_call(id , customer_id,tenant_id,caller_id,type,api_name,count,sharding,created,updated) values(?,?,?,?,?,?,?,?,?,?)"
	GET_APICALL_COUNT  = "select sum(count) as count from api_call WHERE  customer_id = ? and  tenant_id = ?  and api_name = ?  "
	GET_CUSTOMER_COUNT = "select sum(count) as count from api_call WHERE  customer_id = ? and api_name = ?  "
	SELECT_EXIST       = "select id from quota  where customer_id = ? and tenant_id=? and  api_name=? and type = ?"
	SELECT_QUOTA       = "select id , customer_id,tenant_id,api_name,type,value,created,updated from quota  where customer_id = ? and tenant_id=? and  api_name=?"
	UPDATE_QUOTA       = "update quota set  value=? , updated =? where customer_id = ? and tenant_id=? and  api_name=? and type = ?"
	INSERT_QUOTA       = "insert into quota(id , customer_id,tenant_id,api_name,type,value,created,updated) values(?,?,?,?,?,?,?,?)"
	SELECT_QUOTA_ALL   = "select id , customer_id,tenant_id,api_name,type,value,created,updated from quota"
	/*配额类型*/
	EXPIRE  = "expire"
	API_CNT = "api_cnt"
)

var QuotaAll = make(map[string]Quota)
var LoadQuotaTime = time.NewTicker(5 * time.Minute)

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

type Quota struct {
	Id         string
	CustomerId string
	TenantId   string
	ApiName    string
	Type       string
	Value      string
	Created    time.Time
	Updated    time.Time
	lock       sync.Mutex
}

//记录api调用次数
func ApiCallStatistics(w http.ResponseWriter, r *http.Request) bool {

	baseRes := baseResponse{OverQuotaApicall, "request is not available"}
	res := map[string]interface{}{"baseResponse": &baseRes}
	resBody := "request is not available"

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		RetPWriteJSON(w, r, res, &resBody, time.Now())
		return false
	}
	//将body数据写回到request
	rr := bytes.NewReader(body)
	bodyRc := ioutil.NopCloser(rr)
	r.Body = bodyRc

	var args map[string]interface{}
	if err := json.Unmarshal(body, &args); err != nil {
		glog.Errorf(" json.Unmarshal failed (%s)", err)
		RetPWriteJSON(w, r, res, &resBody, time.Now())
		return false
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
			baseRes.ErrMsg = "Auth failure"
			RetPWriteJSON(w, r, res, &resBody, time.Now())
			return false
		}

		tenantId = user.TenantId
		cllerId = user.Uid
	} else { //应用校验
		application, err := getApplicationByToken(token)
		if nil != err || nil == application {
			glog.V(5).Infof("api_call error: [Logon failure]")
			baseRes.ErrMsg = "Auth failure"
			RetPWriteJSON(w, r, res, &resBody, time.Now())
			return false
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
		baseRes.ErrMsg = "not found tenant"
		RetPWriteJSON(w, r, res, &resBody, time.Now())
		return false
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
	//检验此调用是否合法
	if !ValidApiCall(apiCall) {
		RetPWriteJSON(w, r, res, &resBody, time.Now())
		return false
	}

	if apiCallExist(apiCall) { //修改
		addApiCount(apiCall)
	} else { //新增
		insertApiCall(apiCall)
	}
	return true
}

//判断该记录是否存在
func apiCallExist(apiCall *ApiCall) bool {

	rows, err := db.MySQL.Query(APICALL_EXIST, apiCall.CustomerId, apiCall.TenantId, apiCall.CallerId, apiCall.Type, apiCall.ApiName, apiCall.Sharding)
	if rows != nil {
		defer rows.Close()
	}
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
func getApiCallCount(customerId, tenantId, apiName string) int {

	count := 0
	var rows *sql.Rows
	var err error
	//查询但customer下的所有租户调用次数
	if tenantId == "*" {
		rows, err = db.MySQL.Query(GET_CUSTOMER_COUNT, customerId, apiName)
	} else {
		rows, err = db.MySQL.Query(GET_APICALL_COUNT, customerId, tenantId, apiName)
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

//记录api调用次数
func (*app) SyncQuota(w http.ResponseWriter, r *http.Request) {
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
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
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
	token := baseReq["token"].(string)

	//应用校验
	_, err = getApplicationByToken(token)
	if nil != err {
		baseRes.Ret = AuthErr
		return
	}

	quotaM := args["quata"].(map[string]interface{})
	quota := &Quota{
		CustomerId: quotaM["customerId"].(string),
		TenantId:   quotaM["tenantId"].(string),
		ApiName:    quotaM["apiName"].(string),
		Type:       quotaM["type"].(string),
		Value:      quotaM["value"].(string),
	}
	retBool := false
	if isExistQuota(quota) {
		retBool = updateQuota(quota)
	} else {
		retBool = insertQuota(quota)
	}

	if !retBool {
		baseRes.Ret = InternalErr
		baseRes.ErrMsg = "sync quota fail !"
		return
	}

}

// 判断配额是否存在.
func isExistQuota(quota *Quota) bool {

	rows, err := db.MySQL.Query(SELECT_EXIST, quota.CustomerId, quota.TenantId, quota.ApiName, quota.Type)
	if rows != nil {
		defer rows.Close()
	}

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

//修改quota
func updateQuota(quota *Quota) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(UPDATE_QUOTA, quota.Value, time.Now().Local(), quota.CustomerId, quota.TenantId, quota.ApiName, quota.Type)
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}

	if err = tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}

	return true
}

//修改quota
func insertQuota(quota *Quota) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(INSERT_QUOTA, uuid.New(), quota.CustomerId, quota.TenantId, quota.ApiName, quota.Type, quota.Value, time.Now().Local(), time.Now().Local())
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}

	if err = tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}

	return true
}

// 获取配额信息.
func GetQuotas(customerId, tenantId, apiName string) ([]Quota, error) {

	quotas := []Quota{}
	rows, err := db.MySQL.Query(SELECT_QUOTA, customerId, tenantId, apiName)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	for rows.Next() {
		quota := Quota{}
		if err := rows.Scan(&quota.Id, &quota.CustomerId, &quota.TenantId, &quota.ApiName, &quota.Type, &quota.Value, &quota.Created, &quota.Created); err != nil {
			glog.Error(err)
			return nil, err
		}
		quotas = append(quotas, quota)
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
		return nil, err
	}
	return quotas, err
}

//定时加载配额信息5分钟加载一次
func LoadQuotaAll() {

	for _ = range LoadQuotaTime.C {
		rows, err := db.MySQL.Query(SELECT_QUOTA_ALL)
		if err != nil {
			glog.Errorf("load quota err [%s]", err)
		}

		QuotaAll = nil
		QuotaAll = make(map[string]Quota)
		key := bytes.Buffer{}
		for rows.Next() {
			quota := Quota{}
			if err := rows.Scan(&quota.Id, &quota.CustomerId, &quota.TenantId, &quota.ApiName, &quota.Type, &quota.Value, &quota.Created, &quota.Created); err != nil {
				glog.Errorf("load quota err [%s]", err)
				break
			}
			key.WriteString(quota.CustomerId)
			key.WriteString(quota.TenantId)
			key.WriteString(quota.ApiName)
			key.WriteString(quota.Type)
			QuotaAll[key.String()] = quota
			key.Reset()
		}

		if err = rows.Err(); err != nil {
			glog.Errorf("load quota err [%s]", err)
		}
		if rows != nil {
			rows.Close()
		}
	}

}

//检验本次调用是否合法
func ValidApiCall(apiCall *ApiCall) bool {
	//校验时间期限
	key := apiCall.CustomerId + apiCall.TenantId + apiCall.ApiName + EXPIRE
	quota, ok := QuotaAll[key]
	//该租户不存在配置侧查询全局配置
	if !ok {
		key := apiCall.CustomerId + "*" + apiCall.ApiName + EXPIRE
		quota, ok = QuotaAll[key]
	}
	if ok {
		validTime, err := time.ParseInLocation("2006/01/02 15:04:05", quota.Value, time.Local)
		if err != nil || validTime.Before(time.Now().Local()) {
			glog.Error(err)
			return false
		}
		//校验api调用计数
		key = apiCall.CustomerId + apiCall.TenantId + apiCall.ApiName + API_CNT
		quota, ok = QuotaAll[key]
		if !ok {
			key = apiCall.CustomerId + "*" + apiCall.ApiName + API_CNT
			quota, ok = QuotaAll[key]
		}
		if ok {
			quotaCount, err := strconv.Atoi(quota.Value)
			if err != nil {
				glog.Error(err)
				return false
			}
			count := getApiCallCount(quota.CustomerId, quota.TenantId, quota.ApiName)
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
