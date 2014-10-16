package app

import (
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/session"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"time"
)

/*
baseRequest: {
    "uid": "",
    "deviceID": "",
    "deviceType": "", // iOS / Android
    "token": ""
}
state:"active"
sessionId:"1111"
*/
func SessionStat(w http.ResponseWriter, r *http.Request) {

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
	deviceType := baseReq["deviceType"].(string)

	/* Token 校验，分为用户校验和应用校验*/
	token := baseReq["token"].(string)
	//用户校验
	if deviceType == DEVICE_TYPE_IOS || deviceType == DEVICE_TYPE_ANDROID {
		user := getUserByToken(token)
		if nil == user {
			baseRes.Ret = AuthErr
			return
		}
	} else { //应用校验
		_, err := getApplicationByToken(token)
		if nil != err {
			baseRes.Ret = AuthErr
			return
		}
	}
	//修改会话状态
	state := args["state"].(string)
	sessionId := args["sessionId"].(string)
	if ("active" == state || "inactive" == state) && len(sessionId) > 0 {
		if !session.SetSessionStat(sessionId, state) {
			baseRes.Ret = NotFound
			baseRes.ErrMsg = "设置会话状态失败！"
			return
		}
	} else {
		baseRes.Ret = ParamErr
		baseRes.ErrMsg = "参数格式错误，state值只能是：active/inactive,会话id不能为空）"
		return
	}
}

/*
baseRequest: {
    "uid": "",
    "deviceID": "",
    "deviceType": "", // iOS / Android
    "token": ""
}
*/
func (*app) GetSession(w http.ResponseWriter, r *http.Request) {
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

	//获取用户会话列表
	userId := baseReq["uid"].(string)
	ret := session.GetSessions(userId, []string{"all"})
	res["memberList"] = ret

	return
}
