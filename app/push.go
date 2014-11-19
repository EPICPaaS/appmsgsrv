package app

import (
	"encoding/json"
	//"fmt"
	"fmt"
	myrpc "github.com/EPICPaaS/appmsgsrv/rpc"
	"github.com/EPICPaaS/appmsgsrv/session"
	apns "github.com/anachronistic/apns"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	userLen := 0
	qunLen := 0
	// 会话分发
	for _, userName := range toUserNames {

		ns, _ := getNames(userName.(string), sessionArgs)
		names = append(names, ns...)
		//收集发送次数
		if strings.HasSuffix(userName.(string), USER_SUFFIX) { // 如果是推人
			userLen += len(ns)
		} else if strings.HasSuffix(userName.(string), QUN_SUFFIX) { // 如果是推群
			qunLen += len(ns)
		}
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

		//1用户为离线状态  2 根据用户ID查询client是否有IOS，有就合并记录到表中等待推送
		s, _ := session.GetSessionsByUserId(name.Id)
		glog.Infof("start apns push , session[%v], UserId[%v]", len(*s), name.Id)
		if len(*s) == 0 {
			resources, _ := GetResourceByTenantId(application.TenantId)
			apnsToken, _ := getApnsToken(name.Id)
			glog.Infof("toUserId[%v],	msg[%v] ,   resources[%v],	 apnsToken[%v]", name.Id, msg, resources, apnsToken)
			go pushAPNS(msg, resources, apnsToken)
		}

		result := push(key, msgBytes, expire)
		if OK != result {
			baseRes.Ret = result

			glog.Errorf("Push message failed [%v]", msg)

			// 推送分发过程中失败不立即返回，继续下一个推送
		}
	}

	go func() {
		//准备pushCnt（推送统计）信息
		pushCnt := &PushCnt{
			TenantId: application.TenantId,
			CallerId: application.Id,
			Type:     "app",
			PushType: USER_SUFFIX,
			Count:    userLen,
		}
		//统计推用户
		if userLen != 0 {
			StatisticsPush(pushCnt)
		}
		//统计群推
		if qunLen != 0 {
			pushCnt.Count = qunLen
			pushCnt.PushType = QUN_SUFFIX
			StatisticsPush(pushCnt)
		}
	}()

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
	deviceType := baseReq["deviceType"].(string)
	var pushType string // 推送类型

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
		pushType = USER_SUFFIX
		m := getUserByUid(fromUserID)

		msg["fromDisplayName"] = m.NickName
	} else if strings.HasSuffix(toUserName, QUN_SUFFIX) { // 如果是推群
		pushType = QUN_SUFFIX
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
	glog.Infof("start apns push , session[%v], UserId[%v]", len(*s), toUserID)
	if len(*s) == 0 {
		resources, _ := GetResourceByTenantId(user.TenantId)
		apnsToken, _ := getApnsToken(toUserID)
		glog.Infof("toUserId[%v],	msg[%v] ,   resources[%v],	 apnsToken[%v]", toUserID, msg, resources, apnsToken)
		go pushAPNS(msg, resources, apnsToken)
	}

	//准备pushCnt（推送统计）信息
	pushCnt := &PushCnt{
		TenantId: user.TenantId,
		CallerId: user.Uid,
		Type:     deviceType,
		PushType: pushType,
	}
	baseRes.Ret = pushSessions(msg, toUserName, sessionArgs, expire, pushCnt)

	res["msgID"] = "msgid"
	res["clientMsgId"] = msg["clientMsgId"]

	return
}

//  网页断推送消息.
//  1. 单推（@user）
//  2. 群推（@qun）
//  3. 组织机构推（部门 @org，单位 @tenant）
func (*appWeb) WebPush(w http.ResponseWriter, r *http.Request) {

	baseRes := baseResponse{OK, ""}
	res := map[string]interface{}{"baseResponse": &baseRes}
	var callback *string
	defer func() {
		// 返回结果格式化
		resJsonStr := ""
		if resJson, err := json.Marshal(res); err != nil {
			baseRes.ErrMsg = err.Error()
			baseRes.Ret = InternalErr
		} else {
			resJsonStr = string(resJson)
		}
		fmt.Fprintln(w, *callback, "(", resJsonStr, ")")
	}()

	var err error
	var msg = make(map[string]interface{})

	//获取请求数据
	r.ParseForm()

	// Token 校验
	token := r.FormValue("baseRequest[token]")
	var pushType string

	user := getUserByToken(token)
	if nil == user {
		baseRes.Ret = AuthErr
		baseRes.ErrMsg = "Authorization failure"
		return
	}

	tmp := r.FormValue("callbackparam")
	callback = &tmp
	msg["clientMsgId"] = r.FormValue("msg[clientMsgId]")
	msg["msgType"], err = strconv.Atoi(r.FormValue("msg[msgType]"))
	if err != nil {
		baseRes.Ret = ParamErr
		baseRes.ErrMsg = "msgType not is int"
		return
	}
	fromUserName := r.FormValue("msg[fromUserName]")
	fromUserID := fromUserName[:strings.Index(fromUserName, "@")]
	toUserName := r.FormValue("msg[toUserName]")
	toUserID := toUserName[:strings.Index(toUserName, "@")]
	sessionArgs := []string{}
	sessions := r.FormValue("sessions")
	if len(sessions) == 0 {
		// 不存在参数的话默认为 all
		sessionArgs = append(sessionArgs, "all")
	} else {
		sessionArgs = strings.Split(sessions, ",")
	}

	if strings.HasSuffix(toUserName, USER_SUFFIX) { // 如果是推人
		pushType = USER_SUFFIX
		m := getUserByUid(fromUserID)
		msg["fromDisplayName"] = m.NickName
		msg["content"] = r.FormValue("msg[content]")
	} else if strings.HasSuffix(toUserName, QUN_SUFFIX) { // 如果是推群
		pushType = QUN_SUFFIX
		m := getUserByUid(fromUserID)
		qun, err := getQunById(toUserID)

		if nil != err {
			baseRes.Ret = InternalErr
			return
		}

		msg["content"] = fromUserName + "|" + m.Name + "|" + m.NickName + "&&" + r.FormValue("msg[content]")
		msg["fromDisplayName"] = qun.Name
		msg["fromUserName"] = toUserName
	} else { // TODO: 组织机构（部门/单位）推送消息体处理

	}

	// 消息过期时间（单位：秒）
	exp := r.FormValue("msg[expire]")
	expire := 600
	if len(exp) > 0 {
		expire, err = strconv.Atoi(exp)
		if err != nil {
			baseRes.Ret = ParamErr
			return
		}
	}

	//1用户为离线状态  2 根据用户ID查询client是否有IOS，有就合并记录到表中等待推送
	s, _ := session.GetSessionsByUserId(toUserID)
	glog.Infof("start apns push , session[%v], UserId[%v]", len(*s), toUserID)
	if len(*s) == 0 {
		resources, _ := GetResourceByTenantId(user.TenantId)
		apnsToken, _ := getApnsToken(toUserID)
		glog.Infof("toUserId[%v],	msg[%v] ,   resources[%v],	 apnsToken[%v]", toUserID, msg, resources, apnsToken)
		go pushAPNS(msg, resources, apnsToken)
	}

	//准备pushCnt（推送统计）信息
	pushCnt := &PushCnt{
		TenantId: user.TenantId,
		CallerId: user.Uid,
		Type:     APPWEB_TYPE,
		PushType: pushType,
	}
	baseRes.Ret = pushSessions(msg, toUserName, sessionArgs, expire, pushCnt)

	res["msgID"] = "msgid"
	res["clientMsgId"] = r.FormValue("msg[clientMsgId]")
}

//推送IOS离线消息
func pushAPNS(msg map[string]interface{}, resources []*Resource, apnsToken []ApnsToken) {

	var host = "gateway.sandbox.push.apple.com:2195"
	if Conf.ApnsType == "product" {
		host = "gateway.push.apple.com:2195"
	}

	var certFile = ""
	var keyFile = ""
	for _, r := range resources {
		if r.Type == "0" {
			certFile = r.Content
		} else if r.Type == "1" {
			keyFile = r.Content
		}
	}

	if certFile == "" || keyFile == "" {
		glog.Errorf("Push message failed. CertFile [%v] or KeyFile[%v] has a error ", certFile, keyFile)
		return
	}

	for _, t := range apnsToken {

		contentMsg := msg["fromDisplayName"].(string) + ":" + msg["content"].(string)

		if len(contentMsg) > 256 {
			contentMsg = substr(contentMsg, 0, 250) + "..."
		}

		payload := apns.NewPayload()
		payload.Alert = contentMsg
		payload.Badge = 0
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
			glog.Errorf("Push message failed. ApnsToken[%v],Content[%v],Error[%v],Host[%v],CertFile [%v], KeyFile[%v]", t.ApnsToken, alert, resp.Error, host, certFile, keyFile)
			// 推送分发过程中失败不立即返回，继续下一个推送
		} else {
			glog.Infof("Push message successed. ApnsToken[%v],Content[%v],Host[%v]", t.ApnsToken, alert, host)
		}
		// TODO: APNs 回调处理

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
func pushSessions(msg map[string]interface{}, toUserName string, sessionArgs []string, expire int, pushCnt *PushCnt) int {
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
	//统计消息推送记录
	pushCnt.Count = len(names)
	go StatisticsPush(pushCnt)

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

		// id@user (i.e. for offline msg)发送离线消息使用
		name := &Name{Id: userId, SessionId: userId /* user_id 作为 session_id */, ActiveSessionIds: activeSessionIds,
			Suffix: USER_SUFFIX}
		names = append(names, name)

		for _, s := range sessions {
			name := &Name{Id: userId, SessionId: s.Id, ActiveSessionIds: activeSessionIds, Suffix: USER_SUFFIX}
			names = append(names, name)
		}
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
