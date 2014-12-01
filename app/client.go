package app

import (
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"time"
)

const (

	// 获取最新的客户端版本
	SelectLatestClientVerByType = "SELECT * FROM `client_version` WHERE `type` = ? ORDER BY `ver_code` DESC LIMIT 1"
	//持久话apns_token
	InsertApnsToken               = "INSERT INTO `apns_token`(`id`,`user_id`,`device_id`,`apns_token`,`created`,`updated`) VALUES(?,?,?,?,?,?)"
	SelectApnsTokenByUserId       = "SELECT `id`,`user_id`,`device_id`,`apns_token`,`created`,`updated` FROM `apns_token` WHERE `user_id`=? "
	SelectApnsTokenByUserIdTokens = "SELECT `id`,`user_id`,`device_id`,`apns_token`,`created`,`updated` FROM `apns_token` WHERE `user_id`=? AND `apns_token`=?"
	DeleteApnsToken               = "DELETE FROM apns_token where apns_token = ?"
	DeleteApnsTokenByDeviceId     = "DELETE FROM apns_token where device_id = ?"
)

// 客户端结构.
//
// 设备登录的时候会记录.
type Client struct {
	Id              string
	UserId          string
	Type            string
	DeviceId        string
	LatestLoginTime time.Time
	Created         time.Time
	Updated         time.Time
}

// 客户端版本结构.
type ClientVersion struct {
	Id                 string    `json:"id"`
	Type               string    `json:"type"`
	VersionCode        int       `json:"versionCode"`
	VersionName        string    `json:"versionName"`
	VersionDescription string    `json:"versionDesc"`
	DownloadURL        string    `json:"url"`
	FileName           string    `json:"fileName"`
	Created            time.Time `json:"created"`
	Updated            time.Time `json:"updated"`
}

// 客户端版本更新消息内嵌对象内容结构.
type ClientVerUpdateObjectContent struct {
	VersionCode int    `json:"versionCode"`
	VersionName string `json:"versionName"`
	URL         string `json:"url"`
	FileName    string `json:"fileName"`
}

// 客户端版本更新消息结构.
type ClientVerUpdateMsg struct {
	MsgType       int                           `json:"msgType"`
	Content       string                        `json:"content"`
	ObjectContent *ClientVerUpdateObjectContent `json:"objectContent"`
}

//apns_token证书
type ApnsToken struct {
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	DeviceId  string    `json:"deviceId"`
	ApnsToken string    `json:"apnsToken"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

// 登录日志.
func (*device) loginLog(client *Client) {
	now := time.Now()

	client.LatestLoginTime = now
	client.Created = now
	client.Updated = now

	sql := "select id, user_id, type, device_id, latest_login_time, created, updated from client where user_id =? and device_id = ?"

	smt, err := db.MySQL.Prepare(sql)
	if smt != nil {
		defer smt.Close()
	} else {
		return
	}

	if err != nil {
		glog.Error(err)

		return
	}

	row, err := smt.Query(client.UserId, client.DeviceId)

	if row != nil {
		defer row.Close()
	} else {
		return
	}

	exists := false
	rec := &Client{}
	for row.Next() {
		exists = true

		err = row.Scan(&rec.Id, &rec.UserId, &rec.Type, &rec.DeviceId, &rec.LatestLoginTime, &rec.Created, &rec.Updated)
		if err != nil {
			glog.Error(err)

			return
		}
	}

	if exists { // 存在则更新
		updateLoginLog(rec)
	} else { // 不存在则插入
		client.Id = uuid.New()

		insertLoginLog(client)
	}
}

/*插入登陆日志*/
func insertLoginLog(client *Client) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return
	}

	_, err = tx.Exec("INSERT INTO `client`(`id`,`user_id`,`type`,`device_id`, `latest_login_time`, `created`,`updated`)"+
		" VALUES (?,?,?,?,?,?,?)",
		client.Id, client.UserId, client.Type, client.DeviceId, client.LatestLoginTime, client.Created, client.Updated)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}

		return
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)
	}
}

/*修改登陆日志*/
func updateLoginLog(client *Client) {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return
	}

	now := time.Now()
	client.Updated = now
	client.LatestLoginTime = now

	_, err = tx.Exec("UPDATE `client` SET  `latest_login_time` = ? , `updated` = ? WHERE  `id` = ?",
		client.LatestLoginTime, client.Updated, client.Id)
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}

		return
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)
	}
}

/*获取用户设备id（deviceID）*/
func getDeviceIds(userId string) []string {
	ret := []string{}

	sql := "select device_id from client where user_id =?"

	smt, err := db.MySQL.Prepare(sql)
	if smt != nil {
		defer smt.Close()
	} else {
		return ret
	}

	if err != nil {
		glog.Error(err)

		return ret
	}

	row, err := smt.Query(userId)

	if row != nil {
		defer row.Close()
	} else {
		return ret
	}

	for row.Next() {
		deviceId := ""

		err = row.Scan(&deviceId)
		if err != nil {
			glog.Error(err)

			return ret
		}

		ret = append(ret, deviceId)
	}

	return ret
}

//移动端检查更新.
func (*device) CheckUpdate(w http.ResponseWriter, r *http.Request) {
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
		baseRes.Ret = ParamErr
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

	// Token 校验
	token := baseReq["token"].(string)
	user := getUserByToken(token)
	if nil == user {
		baseRes.Ret = AuthErr

		return
	}

	deviceType := args["type"].(string)

	clientVersion, err := getLatestVerion(deviceType)
	if nil != err {
		baseRes.Ret = InternalErr

		return
	}

	// 组装返回结果
	objectContent := ClientVerUpdateObjectContent{
		VersionCode: clientVersion.VersionCode, VersionName: clientVersion.VersionName, URL: clientVersion.DownloadURL,
		FileName: clientVersion.FileName}
	msg := ClientVerUpdateMsg{MsgType: 1001, Content: clientVersion.VersionDescription, ObjectContent: &objectContent}

	res["msg"] = msg
}

/*持久话apns_token*/
func (*device) AddApnsToken(w http.ResponseWriter, r *http.Request) {

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
		baseRes.Ret = ParamErr
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
	// Token 校验
	token := baseReq["token"].(string)
	user := getUserByToken(token)
	if nil == user {
		baseRes.Ret = AuthErr
		baseRes.ErrMsg = "Authentication failed"
		return
	}

	apnsTokenStr := args["apns_token"].(string)
	deviceId := baseReq["deviceID"].(string)
	if len(apnsTokenStr) == 0 || len(deviceId) == 0 {
		baseRes.Ret = ParamErr
		return
	}

	apnsToken := &ApnsToken{
		UserId:    user.Uid,
		DeviceId:  deviceId,
		ApnsToken: apnsTokenStr,
		Created:   time.Now().Local(),
		Updated:   time.Now().Local(),
	}
	//先删除该设备对应的信息
	deleteApnsTokenByDeviceId(deviceId)
	//再插入该设备对应信息
	if insertApnsToken(apnsToken) {
		baseRes.Ret = OK
		baseRes.ErrMsg = "save apns_token success"
		return
	} else {
		baseRes.Ret = InternalErr
		baseRes.ErrMsg = "Sava apns_token faild"
		return
	}
}

// 在数据库中查询指定类型客户端的最新的版本.
func getLatestVerion(deviceType string) (*ClientVersion, error) {
	row := db.MySQL.QueryRow(SelectLatestClientVerByType, deviceType)

	clientVer := ClientVersion{}
	if err := row.Scan(&clientVer.Id, &clientVer.Type, &clientVer.VersionCode, &clientVer.VersionName,
		&clientVer.VersionDescription, &clientVer.DownloadURL, &clientVer.FileName, &clientVer.Created,
		&clientVer.Updated); err != nil {
		glog.Error(err)

		return nil, err
	}

	return &clientVer, nil
}

/*保存apnstoken证书*/
func insertApnsToken(apnsToken *ApnsToken) bool {

	rows, err := db.MySQL.Query(SelectApnsTokenByUserIdTokens, apnsToken.UserId, apnsToken.ApnsToken)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		glog.Error(err)
		return false
	}
	if !rows.Next() { //不存在记录才添加
		tx, err := db.MySQL.Begin()
		if err != nil {
			glog.Error(err)
			return false
		}

		_, err = tx.Exec(InsertApnsToken, uuid.New(), apnsToken.UserId, apnsToken.DeviceId, apnsToken.ApnsToken, apnsToken.Created, apnsToken.Updated)
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
	}
	return true
}

/*根据userd获取apnsToken证书*/
func getApnsToken(userId string) ([]ApnsToken, error) {

	ret := []ApnsToken{}
	glog.Infoln("userId", userId)
	rows, err := db.MySQL.Query(SelectApnsTokenByUserId, userId)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		apnsToken := ApnsToken{}
		if err := rows.Scan(&apnsToken.Id, &apnsToken.UserId, &apnsToken.DeviceId, &apnsToken.ApnsToken, &apnsToken.Created, &apnsToken.Updated); err != nil {
			glog.Error(err)
			return nil, err
		}
		ret = append(ret, apnsToken)
	}

	if err := rows.Err(); err != nil {
		glog.Error(err)
		return nil, err
	}
	return ret, nil
}

//根据apns_token删除token
func deleteApnsToken(apns_token string) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}

	_, err = tx.Exec(DeleteApnsToken, apns_token)
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

//根据apns_token删除token
func deleteApnsTokenByDeviceId(deviceId string) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}

	_, err = tx.Exec(DeleteApnsTokenByDeviceId, deviceId)
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
