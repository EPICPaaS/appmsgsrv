package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
)

const (

	// 获取最新的客户端版本
	SelectLatestClientVerByType = "SELECT * FROM `client_version` WHERE `type` = ? ORDER BY `ver_code` DESC LIMIT 1"
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

// 移动端检查更新.
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
