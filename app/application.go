package app

import (
	//"database/sql"
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	// 根据 id 查询应用记录.
	SelectApplicationById = "SELECT  * FROM `application` WHERE `id` = ?"
	// 查询应用记录.
	SelectAllApplication = "SELECT `id`, `name`, `name`,`status`, `sort`,`avatar`, `tenant_id`, `name_py`, `name_quanpin` FROM `application`"
	// 根据 token 获取应用记录.
	SelectApplicationByToken = "SELECT * FROM `application` WHERE `token` = ?"
	//根据应用ID查询应用操作项列表
	SelectAppOpertionByAppId = "SELECT  `id`, `app_id`, `content`,`action`, `msg_type`,`sort`   FROM `operation` WHERE `app_id` = ?  and  parent_id  is  null  order by  sort "
	//根据操作项父ID查询应用操作项列表
	SelectAppOpertionByParentId = "SELECT `id`, `app_id`, `content`,`action`, `msg_type`,`sort`  FROM `operation` WHERE `parent_id` = ?  order by  sort "
)

// 应用结构.
type application struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	Type      string    `json:"type"`
	Status    int       `json:"status"`
	Sort      int       `json:"sort"`
	Level     int       `json:"level"`
	Avatar    string    `json:"avatar"`
	TenantId  string    `json:tenantId`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	PYInitial string    `json:"pYInitial"`
	PYQuanPin string    `json:"pYQuanPin"`
}

// 应用操作项
type operation struct {
	Id      string `json:"id"`
	AppId   string `json:"appId"`
	Content string `json:"content"`
	Action  string `json:"action"`
	MsgType string `json:"msgType"`
	Sort    int    `json:"sort"`

	OpertionList []*operation `json:"operationList"`
}

// 根据 id 查询应用记录.
func getApplication(appId string) (*application, error) {
	row := db.MySQL.QueryRow(SelectApplicationById, appId)

	application := application{}

	if err := row.Scan(&application.Id, &application.Name, &application.Token, &application.Type, &application.Status,
		&application.Sort, &application.Level, &application.Avatar, &application.TenantId, &application.Created, &application.Updated, &application.PYInitial, &application.PYQuanPin); err != nil {
		glog.Error(err)

		return nil, err
	}

	return &application, nil
}

func getAllApplication() ([]*member, error) {
	rows, _ := db.MySQL.Query(SelectAllApplication)
	if rows != nil {
		defer rows.Close()
	}
	ret := []*member{}
	for rows.Next() {
		rec := member{}

		if err := rows.Scan(&rec.Uid, &rec.Name, &rec.NickName, &rec.Status, &rec.Sort, &rec.Avatar, &rec.TenantId, &rec.PYInitial, &rec.PYQuanPin); err != nil {
			glog.Error(err)

			return nil, err
		}

		rec.UserName = rec.Uid + APP_SUFFIX
		ret = append(ret, &rec)

	}

	return ret, nil
}

//根据appId获取应用的列表项
func getAppOpertionListByAppId(appId string) ([]*operation, error) {
	rows, _ := db.MySQL.Query(SelectAppOpertionByAppId, appId)
	if rows != nil {
		defer rows.Close()
	}
	ret := []*operation{}
	for rows.Next() {
		rec := operation{}

		if err := rows.Scan(&rec.Id, &rec.AppId, &rec.Content, &rec.Action, &rec.MsgType, &rec.Sort); err != nil {
			glog.Error(err)
			return nil, err
		}
		crows, _ := db.MySQL.Query(SelectAppOpertionByParentId, &rec.Id)
		if crows != nil {
			defer crows.Close()
		}
		opertionList := []*operation{}
		for crows.Next() {
			crec := operation{}

			if err := crows.Scan(&crec.Id, &crec.AppId, &crec.Content, &crec.Action, &crec.MsgType, &crec.Sort); err != nil {
				glog.Error(err)
				return nil, err
			}
			opertionList = append(opertionList, &crec)
		}
		rec.OpertionList = opertionList
		ret = append(ret, &rec)
	}

	return ret, nil

}

// 根据 token 查询应用记录.
func getApplicationByToken(token string) (*application, error) {
	row := db.MySQL.QueryRow(SelectApplicationByToken, token)

	application := application{}

	if err := row.Scan(&application.Id, &application.Name, &application.Token, &application.Type, &application.Status,
		&application.Sort, &application.Level, &application.Avatar, &application.TenantId, &application.Created, &application.Updated, &application.PYInitial, &application.PYQuanPin); err != nil {
		glog.Error(err)

		return nil, err
	}

	return &application, nil
}

/*
*   根据Application获取Member
 */

func (*device) GetApplicationList(w http.ResponseWriter, r *http.Request) {
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
		baseRes.ErrMsg = "会话超时请重新登录"
		return
	}

	members, err := getAllApplication()
	if err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr
		return
	}

	res["memberList"] = members
	res["memberCount"] = len(members)
}

/*
* 根据应用信息获取应用的操作项
**/
func (*device) GetAppOperationList(w http.ResponseWriter, r *http.Request) {
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
		baseRes.ErrMsg = "会话超时请重新登录"
		return
	}

	username := args["username"].(string)
	if !strings.HasSuffix(username, APP_SUFFIX) {
		baseRes.Ret = AuthErr
		baseRes.ErrMsg = "username没有以 " + APP_SUFFIX + " 结尾"
		return
	}

	appId := strings.Split(username, APP_SUFFIX)[0]
	opertions, err := getAppOpertionListByAppId(appId)
	if err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr
		return
	}

	res["operationList"] = opertions
	res["operationCount"] = len(opertions)
}
