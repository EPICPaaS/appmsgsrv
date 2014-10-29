package app

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	myrpc "github.com/EPICPaaS/appmsgsrv/rpc"
	"github.com/EPICPaaS/appmsgsrv/session"
	apns "github.com/anachronistic/apns"
	"github.com/golang/glog"
)

// 推送 Name.
type Name struct {
	Id               string
	SessionId        string
	ActiveSessionIds []string
	Suffix           string
}

// 转换为推送 key.
func (n *Name) toKey() string {
	return n.SessionId + n.Suffix
}

// 应用端推送消息给用户.
func (*app) UserPush(w http.ResponseWriter, r *http.Request) {
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
	application, err := getApplicationByToken(token)
	if nil != err {
		baseRes.Ret = AuthErr

		return
	}

	msg := map[string]interface{}{}

	content := args["content"].(string)
	msg["content"] = content
	msg["msgType"] = args["msgType"].(string)
	msg["objectContent"] = args["objectContent"]

	toUserNames := args["toUserNames"].([]interface{})
	if len(toUserNames) > 1000 { // 一次最多只能推送 1000 人
		baseRes.Ret = TooLong

		return
	}

	sessionArgs := []string{}
	_, exists := args["sessions"]
	if !exists {
		// 不存在参数的话默认为 all
		sessionArgs = append(sessionArgs, "all")
	} else {
		for _, arg := range args["sessions"].([]interface{}) {
			sessionArgs = append(sessionArgs, arg.(string))
		}
	}

	appId := args["objectContent"].(map[string]interface{})["appId"].(string)

	msg["fromUserName"] = appId + APP_SUFFIX
	msg["fromDisplayName"] = application.Name

	// 消息过期时间（单位：秒）
	exp := args["expire"]
	expire := 600
	if nil != exp {
		expire = int(exp.(float64))
	}

	names := []*Name{}
	// 会话分发
	for _, userName := range toUserNames {
		ns, _ := getNames(userName.(string), sessionArgs)

		names = append(names, ns...)
	}

	// 推送
	for _, name := range names {
		key := name.toKey()

		msg["toUserName"] = name.Id + name.Suffix
		msg["toUserKey"] = key

		msg["activeSessions"] = name.ActiveSessionIds

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			baseRes.Ret = ParamErr
			glog.Error(err)

			return
		}

		result := push(key, msgBytes, expire)
		if OK != result {
			baseRes.Ret = result

			glog.Errorf("Push message failed [%v]", msg)

			// 推送分发过程中失败不立即返回，继续下一个推送
		}
	}

	return
}

// 客户端设备推送消息.
//
//  1. 单推（@user）
//  2. 群推（@qun）
//  3. 组织机构推（部门 @org，单位 @tenant）
func (*device) Push(w http.ResponseWriter, r *http.Request) {
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

	msg := args["msg"].(map[string]interface{})
	fromUserName := msg["fromUserName"].(string)
	fromUserID := fromUserName[:strings.Index(fromUserName, "@")]
	toUserName := msg["toUserName"].(string)
	toUserID := toUserName[:strings.Index(toUserName, "@")]
	sessionArgs := []string{}
	_, exists := args["sessions"]
	if !exists {
		// 不存在参数的话默认为 all
		sessionArgs = append(sessionArgs, "all")
	} else {
		for _, arg := range args["sessions"].([]interface{}) {
			sessionArgs = append(sessionArgs, arg.(string))
		}
	}

	if strings.HasSuffix(toUserName, USER_SUFFIX) { // 如果是推人
		m := getUserByUid(fromUserID)

		msg["fromDisplayName"] = m.NickName
	} else if strings.HasSuffix(toUserName, QUN_SUFFIX) { // 如果是推群
		m := getUserByUid(fromUserID)

		qun, err := getQunById(toUserID)

		if nil != err {
			baseRes.Ret = InternalErr

			return
		}

		msg["content"] = fromUserName + "|" + m.Name + "|" + m.NickName + "&&" + msg["content"].(string)
		msg["fromDisplayName"] = qun.Name
		msg["fromUserName"] = toUserName
	} else { // TODO: 组织机构（部门/单位）推送消息体处理

	}

	// 消息过期时间（单位：秒）
	exp := msg["expire"]
	expire := 600
	if nil != exp {
		expire = int(exp.(float64))
	}
	//1用户为离线状态  2 根据用户ID查询client是否有IOS，有就合并记录到表中等待推送
	s, _ := session.GetSessionsByUserId(toUserID)
	if nil == s {
		resources, _ := GetResourceByTenantId(user.TenantId)
		apnsToken, _ := getApnsToken(toUserID)
		go pushAPNS(msg, resources, apnsToken)

	}
	baseRes.Ret = pushSessions(msg, toUserName, sessionArgs, expire)

	res["msgID"] = "msgid"
	res["clientMsgId"] = msg["clientMsgId"]

	return
}

//推送IOS离线消息
func pushAPNS(msg map[string]interface{}, resources []*Resource, apnsToken []ApnsToken) {

	var host = "gateway.sandbox.push.apple.com:2195"
	var certFile = ""
	var keyFile = ""

	for _, r := range resources {
		if r.Type == "0" {
			certFile = r.Content
		} else if r.Type == "1" {
			keyFile = r.Content
		}
	}

	for _, t := range apnsToken {

		contentMsg := msg["fromDisplayName"].(string) + ":" + msg["content"].(string)

		if len(contentMsg) > 256 {
			contentMsg = substr(contentMsg, 0, 250) + "..."
		}

		payload := apns.NewPayload()
		payload.Alert = contentMsg
		payload.Badge = 42
		payload.Sound = "bingbong.aiff"
		payload.Category = "Test!"
		payload.ContentAvailable = 1

		pn := apns.NewPushNotification()
		pn.DeviceToken = t.ApnsToken
		pn.AddPayload(payload)
		if nil != msg["customFilelds"] {
			customFilelds := msg["customFilelds"].(map[string]interface{})
			for key, value := range customFilelds {
				pn.Set(key, value)
			}

		}

		client := apns.NewClient(host, certFile, keyFile)
		resp := client.Send(pn)
		alert, _ := pn.PayloadString()
		if !resp.Success {
			glog.Errorf("Push message failed [%v],Error[%v]", alert, resp.Error)
			// 推送分发过程中失败不立即返回，继续下一个推送
		}

	}
}

func substr(s string, pos, length int) string {
	bytes := []byte(s)
	l := pos + length
	if l > len(bytes) {
		l = len(bytes)
	}
	return string(bytes[pos:l])
}

// 按会话推送.
func pushSessions(msg map[string]interface{}, toUserName string, sessionArgs []string, expire int) int {
	names, _ := getNames(toUserName, sessionArgs)

	// 推送
	for _, name := range names {
		key := name.toKey()

		msg["toUserKey"] = key
		msg["toUserName"] = name.Id + name.Suffix
		msg["activeSessions"] = name.ActiveSessionIds

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			glog.Error(err)

			return ParamErr
		}

		result := push(key, msgBytes, expire)
		if OK != result {
			glog.Errorf("Push message failed [%v]", msg)

			// 推送分发过程中失败不立即返回，继续下一个推送
		}
	}

	return OK
}

// 按 key 推送.
func push(key string, msgBytes []byte, expire int) int {
	node := myrpc.GetComet(key)

	if node == nil || node.CometRPC == nil {
		glog.Errorf("Get comet node failed [key=%s]", key)

		return NotFoundServer
	}

	client := node.CometRPC.Get()
	if client == nil {
		glog.Errorf("Get comet node RPC client failed [key=%s]", key)

		return NotFoundServer
	}

	pushArgs := &myrpc.CometPushPrivateArgs{Msg: json.RawMessage(msgBytes), Expire: uint(expire), Key: key}

	ret := OK
	if err := client.Call(myrpc.CometServicePushPrivate, pushArgs, &ret); err != nil {
		glog.Errorf("client.Call(\"%s\", \"%v\", &ret) error(%v)", myrpc.CometServicePushPrivate, string(msgBytes), err)

		return InternalErr
	}

	glog.V(3).Infof("Pushed a message to [key=%s]", key)

	return ret
}

// 构造推送 name 集.
func buildNames(userIds []string, sessionArgs []string) (names []*Name) {
	for _, userId := range userIds {
		sessions := session.GetSessions(userId, sessionArgs)

		activeSessionIds := []string{}
		for _, s := range sessions {
			if session.SESSION_STATE_ACTIVE == s.State {
				activeSessionIds = append(activeSessionIds, s.Id)
			}
		}

		for _, s := range sessions {
			name := &Name{Id: userId, SessionId: s.Id, ActiveSessionIds: activeSessionIds, Suffix: USER_SUFFIX}
			names = append(names, name)
		}

		// id@user (i.e. for offline msg)
		name := &Name{Id: userId, SessionId: userId /* user_id 作为 session_id */, ActiveSessionIds: activeSessionIds,
			Suffix: USER_SUFFIX}
		names = append(names, name)
	}

	return names
}

// 根据 toUserName 获得最终推送的 name 集（包含会话分发处理）.
func getNames(toUserName string, sessionArgs []string) (names []*Name, pushType string) {
	toUserId := toUserName[:strings.Index(toUserName, "@")]

	if strings.HasSuffix(toUserName, QUN_SUFFIX) { // 群推
		qunId := toUserName[:len(toUserName)-len(QUN_SUFFIX)]

		userIds, err := getUserIdsInQun(qunId)

		if nil != err {
			return names, QUN_SUFFIX
		}

		names = buildNames(userIds, sessionArgs)

		return names, QUN_SUFFIX
	} else if strings.HasSuffix(toUserName, ORG_SUFFIX) { // 组织机构部门推
		orgId := toUserName[:len(toUserName)-len(ORG_SUFFIX)]

		users := getUserListByOrgId(orgId)

		if nil == users {
			return names, ORG_SUFFIX
		}

		userIds := []string{}
		for _, user := range users {
			userIds = append(userIds, user.Uid)
		}

		names = buildNames(userIds, sessionArgs)

		return names, ORG_SUFFIX
	} else if strings.HasSuffix(toUserName, TENANT_SUFFIX) { // 组织机构单位推
		tenantId := toUserName[:len(toUserName)-len(TENANT_SUFFIX)]

		users := getUserListByTenantId(tenantId)

		if nil == users {
			return names, TENANT_SUFFIX
		}

		userIds := []string{}
		for _, user := range users {
			userIds = append(userIds, user.Uid)
		}

		names = buildNames(userIds, sessionArgs)

		return names, TENANT_SUFFIX
	} else if strings.HasSuffix(toUserName, USER_SUFFIX) { // 用户推
		userIds := []string{toUserId}

		names = buildNames(userIds, sessionArgs)

		return names, USER_SUFFIX
	} else if strings.HasSuffix(toUserName, APP_SUFFIX) { // 应用推
		glog.Warningf("应用推需要走单独的接口")
		return names, APP_SUFFIX
	} else {
		return names, "@UNDEFINDED"
	}
}
