package app

import (
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"time"
)

const (

	// 获取最新的客户端版本
	SelectLatestClientVerByType = "SELECT * FROM `client_version` WHERE `type` = ? ORDER BY `ver_code` DESC LIMIT 1"
)

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
