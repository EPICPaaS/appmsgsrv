package app

import (
	"encoding/json"
	"fmt"
	myrpc "github.com/EPICPaaS/appmsgsrv/rpc"
	"github.com/EPICPaaS/appmsgsrv/session"
	apns "github.com/anachronistic/apns"
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
	DisplayName      string
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
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
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
	msg["msgType"] = args["msgType"].(float64)
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
	expire := Conf.MsgExpire
	if nil != exp {
		expire = int(exp.(float64))
	}

	names := []*Name{}

	//准备pushCnt（推送统计）信息
	tenant := getTenantById(application.TenantId)
	if tenant == nil {
		baseRes.Ret = InternalErr
		return
	}
	pushCnt := &PushCnt{
		CustomerId: tenant.CustomerId,
		TenantId:   application.TenantId,
		CallerId:   application.Id,
		Type:       "app",
	}

	//检验是否可以发送消息
	if !ValidPush(pushCnt) {
		baseRes.Ret = OverQuotaPush
		return
	}

	userLen := 0
	qunLen := 0
	appLen := 0

	// 会话分发
	for _, userName := range toUserNames {

		if strings.HasSuffix(userName.(string), APP_SUFFIX) { //推消息给应用,就直接推送

			msg["toUserName"] = userName.(string)
			appId := userName.(string)[:strings.Index(userName.(string), "@")]
			application, err = getApplication(appId) //查找
			if err != nil {
				continue
			}
			msg["toDisplayName"] = application.Name
			msg["toUserKey"] = userName.(string)

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				baseRes.Ret = ParamErr
				logger.Error(err)
				return
			}

			push(userName.(string), msgBytes, expire)
			appLen++
			continue
		}

		ns, _ := getNames(userName.(string), sessionArgs)
		names = append(names, ns...)
		//收集发送次数
		if strings.HasSuffix(userName.(string), USER_SUFFIX) { // 如果是推人
			userLen += len(ns)
		} else if strings.HasSuffix(userName.(string), QUN_SUFFIX) { // 如果是推群
			qunLen += len(ns)
		}
	}

	/*记录发送文件信息*/
	msgType, ok := msg["msgType"].(float64)
	if ok && msgType == 2 {
		objectContent, ok := msg["objectContent"].(map[string]interface{})
		if !ok {
			baseRes.Ret = ParamErr
			return
		}
		responseUpload, ok := objectContent["responseUpload"].(map[string]interface{})
		if !ok {
			baseRes.Ret = ParamErr
			return
		}

		fileId := responseUpload["fid"].(string)
		fileName := responseUpload["fileName"].(string)
		fileUrl := responseUpload["fileUrl"].(string)
		size := int(responseUpload["size"].(float64))
		fileLink := &FileLink{
			SenderId: application.Id,
			FileId:   fileId,
			FileName: fileName,
			FileUrl:  fileUrl,
			Size:     size,
		}
		go SaveFileLinK(fileLink)
	}

	shieldMsgs := GetShieldTargetMsg(application.Id, 1)
	// 推送
	for _, name := range names {

		//判断该用户是否屏蔽里该消息
		isShield := false
		for _, s := range shieldMsgs {
			if s.SourceId == name.Id {
				isShield = true
				break
			}
		}

		if isShield {
			continue
		}

		key := name.toKey()

		msg["toUserName"] = name.Id + name.Suffix
		msg["toDisplayName"] = name.DisplayName
		msg["toUserKey"] = key

		msg["activeSessions"] = name.ActiveSessionIds

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			baseRes.Ret = ParamErr
			logger.Error(err)

			return
		}

		go pushWithAPNS(application.TenantId, name.Id, msg)

		logger.Infof("toKey [%v] ", key)

		//TODO 应用推送消息忽略mid
		result, _ := push(key, msgBytes, expire)
		if OK != result {
			baseRes.Ret = result

			logger.Errorf("Push message failed [%v]", msg)

			// 推送分发过程中失败不立即返回，继续下一个推送
		}
	}

	go func() {
		//统计推用户
		if userLen != 0 {
			pushCnt.Count = userLen
			pushCnt.PushType = USER_SUFFIX
			StatisticsPush(pushCnt)
		}
		//统计群推
		if qunLen != 0 {
			pushCnt.Count = qunLen
			pushCnt.PushType = QUN_SUFFIX
			StatisticsPush(pushCnt)
		}
		//统计推应用
		if appLen != 0 {
			pushCnt.Count = appLen
			pushCnt.PushType = APP_SUFFIX
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
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
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
	/*deviceID 用于屏蔽当前设备和过滤当前客户端离线消息*/
	msg["deviceID"] = baseReq["deviceID"].(string)
	// 消息过期时间（单位：秒）
	exp := msg["expire"]
	expire := Conf.MsgExpire
	if nil != exp {
		expire = int(exp.(float64))
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

	if strings.HasSuffix(toUserName, USER_SUFFIX) { // 如果是推人
		pushType = USER_SUFFIX
		m := getUserByUid(fromUserID)
		msg["fromDisplayName"] = m.NickName

		go pushWithAPNS(user.TenantId, toUserID, msg)

	} else if strings.HasSuffix(toUserName, QUN_SUFFIX) { // 如果是推群
		pushType = QUN_SUFFIX
		m := getUserByUid(fromUserID)

		qun, err := getQunById(toUserID)

		if nil != err {
			baseRes.Ret = InternalErr
			return
		}

		/*校验该用户是否存在于该群*/
		qunUserIds, err := getUserIdsInQun(toUserID)
		if nil != err {
			baseRes.Ret = InternalErr
			return
		}
		isQunUser := false
		for _, userId := range qunUserIds {
			if userId == fromUserID {
				isQunUser = true
				break
			}
		}
		if !isQunUser {
			baseRes.Ret = DeleteUser
			return
		}
		msg["content"] = fromUserName + "|" + m.Name + "|" + m.NickName + "&&" + msg["content"].(string)
		msg["fromDisplayName"] = qun.Name
		msg["fromUserName"] = toUserName

		go func() {
			logger.Infof("qunUserIds[%v]", qunUserIds)
			for _, usid := range qunUserIds {
				logger.Infof("usid[%v]", usid)
				pushWithAPNS(user.TenantId, usid, msg)
			}
		}()

	} else if strings.HasSuffix(toUserName, APP_SUFFIX) {

		pushType = APP_SUFFIX
		m := getUserByUid(fromUserID)
		msg["fromDisplayName"] = m.NickName

	} else { // TODO: 组织机构（部门/单位）推送消息体处理

	}

	//准备pushCnt（推送统计）信息
	tenant := getTenantById(user.TenantId)
	pushCnt := PushCnt{
		TenantId: user.TenantId,
		CallerId: user.Uid,
		Type:     deviceType,
		PushType: pushType,
	}
	/*不属于任何组织结构的用户就不统计配额*/
	if tenant != nil {
		pushCnt.CustomerId = tenant.CustomerId
	}
	var mid int64 = 0
	baseRes.Ret, mid = pushSessions(msg, toUserName, sessionArgs, expire, pushCnt)

	res["msgID"] = "msgid"
	res["clientMsgId"] = msg["clientMsgId"]
	res["mid"] = mid
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
	msg["deviceID"] = r.FormValue("baseRequest[deviceID]")
	msg["fromUserName"] = r.FormValue("msg[fromUserName]")
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

		go pushWithAPNS(user.TenantId, toUserID, msg)

	} else if strings.HasSuffix(toUserName, QUN_SUFFIX) { // 如果是推群
		pushType = QUN_SUFFIX
		m := getUserByUid(fromUserID)
		qun, err := getQunById(toUserID)

		if nil != err {
			baseRes.Ret = InternalErr
			return
		}

		/*校验该用户是否存在于该群*/
		qunUserIds, err := getUserIdsInQun(toUserID)
		if nil != err {
			baseRes.Ret = InternalErr
			return
		}
		isQunUser := false
		for _, userId := range qunUserIds {
			if userId == fromUserID {
				isQunUser = true
				break
			}
		}
		if !isQunUser {
			baseRes.Ret = DeleteUser
			return
		}

		msg["content"] = fromUserName + "|" + m.Name + "|" + m.NickName + "&&" + r.FormValue("msg[content]")
		msg["fromDisplayName"] = qun.Name
		msg["fromUserName"] = toUserName

		go func() {
			logger.Infof("qunUserIds[%v]", qunUserIds)
			for _, usid := range qunUserIds {
				logger.Infof("usid[%v]", usid)
				pushWithAPNS(user.TenantId, usid, msg)
			}
		}()

	} else { // TODO: 组织机构（部门/单位）推送消息体处理

	}

	// 消息过期时间（单位：秒）
	exp := r.FormValue("msg[expire]")
	expire := Conf.MsgExpire
	if len(exp) > 0 {
		expire, err = strconv.Atoi(exp)
		if err != nil {
			baseRes.Ret = ParamErr
			return
		}
	}

	//准备pushCnt（推送统计）信息
	//获取租户信息
	tenant := getTenantById(user.TenantId)
	pushCnt := PushCnt{
		TenantId: user.TenantId,
		CallerId: user.Uid,
		Type:     APPWEB_TYPE,
		PushType: pushType,
	}
	if tenant != nil {
		pushCnt.CustomerId = tenant.CustomerId
	}
	var mid int64 = 0
	baseRes.Ret, mid = pushSessions(msg, toUserName, sessionArgs, expire, pushCnt)

	res["msgID"] = "msgid"
	res["clientMsgId"] = r.FormValue("msg[clientMsgId]")
	res["mid"] = mid
}

//apns消息封装
func pushWithAPNS(tenantId, userId string, msg map[string]interface{}) {
	//1用户为离线状态的设备（多个设备只发离线的）  2 根据用户ID查询client是否有IOS，有就合并记录到表中等待推送
	apnsToken, _ := getApnsToken(userId)
	logger.Infof("apnsToken[%v], UserId[%v]", apnsToken, userId)
	if len(apnsToken) > 0 {
		s, _ := session.GetSessionsByUserId(userId)
		// sessionId的规则是   用户ID_设备类型-设备号     9B706D583BF450D82127C7117F821B9F_iOS-b3afa122e562d91b39e330a11ad

		//var device_type = ""
		var device_id = ""
		for _, t := range *s {
			logger.Infof("%v", t.Id)
			parts := strings.Split(t.Id, "_")
			if len(parts) > 1 {
				tmp := strings.Split(parts[1], "-")
				//device_type = tmp[0]
				device_id = tmp[1]
				//排除在线的设备
				for k, v := range apnsToken {
					if v.DeviceId == device_id {
						kk := k + 1
						apnsToken = append(apnsToken[:k], apnsToken[kk:]...)
					}
				}

			}
		}

		resources, _ := GetResourceByTenantId(tenantId)
		logger.Infof("toUserId[%v], msg[%v] ,  resources[%v],	 apnsToken[%v]", userId, msg, resources, apnsToken)
		pushAPNS(msg, resources, apnsToken)
	}
	/**
	s, _ := session.GetSessionsByUserId(userId)
	logger.Infof("start apns push , session[%v], UserId[%v]", len(*s), userId)
	if len(*s) == 0 {
		apnsToken, _ := getApnsToken(userId)
		logger.Infof("apnsToken[%v], UserId[%v]", apnsToken, userId)
		if len(apnsToken) > 0 {

			resources, _ := GetResourceByTenantId(tenantId)
			logger.Infof("toUserId[%v], msg[%v] ,  resources[%v],	 apnsToken[%v]", userId, msg, resources, apnsToken)
			pushAPNS(msg, resources, apnsToken)
		}
	}
	**/
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
		logger.Errorf("Push message failed. CertFile [%v] or KeyFile[%v] has a error ", certFile, keyFile)
		return
	}

	for _, t := range apnsToken {
		content := msg["content"].(string)
		toUserName := msg["toUserName"].(string)

		if strings.HasSuffix(toUserName, QUN_SUFFIX) {
			if strings.Contains(content, "&&") {
				content = strings.Split(content, "&&")[1]
			}
		}
		contentMsg := msg["fromDisplayName"].(string) + ":" + content
		logger.Infof(" contentMsg[%v] ", contentMsg)

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
			logger.Errorf("Push message failed. ApnsToken[%v],Content[%v],Error[%v],Host[%v],CertFile [%v], KeyFile[%v]", t.ApnsToken, alert, resp.Error, host, certFile, keyFile)
			// 推送分发过程中失败不立即返回，继续下一个推送

			//只删除失效类型
			if resp.Error.Error() == apns.ApplePushResponses[8] || resp.Error.Error() == apns.ApplePushResponses[5] {
				if deleteApnsToken(t.ApnsToken) {
					logger.Trace("delete INVALID_TOKEN  succeed")
				} else {
					logger.Trace("delete  INVALID_TOKEN failure")
				}
			}

		} else {
			logger.Infof("Push message successed. ApnsToken[%v],Content[%v],Host[%v]", t.ApnsToken, alert, host)
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
func pushSessions(msg map[string]interface{}, toUserName string, sessionArgs []string, expire int, pushCnt PushCnt) (int, int64) {

	//不属于任何组织机构的不统计配额不限制
	if len(pushCnt.TenantId) != 0 && !ValidPush(&pushCnt) {
		return OverQuotaPush, 0
	}

	/*记录发送文件信息*/
	msgType, ok := msg["msgType"].(float64)
	if ok && msgType == 2 {
		objectContent, ok := msg["objectContent"].(map[string]interface{})
		if !ok {
			return ParamErr, 0
		}
		responseUpload, ok := objectContent["responseUpload"].(map[string]interface{})
		if !ok {
			return ParamErr, 0
		}

		fileId := responseUpload["fid"].(string)
		fileName := responseUpload["fileName"].(string)
		fileUrl := responseUpload["fileUrl"].(string)
		size := int(responseUpload["size"].(float64))
		fileLink := &FileLink{
			SenderId: pushCnt.CallerId, // pushCnt.CallerId为发送者的id
			FileId:   fileId,
			FileName: fileName,
			FileUrl:  fileUrl,
			Size:     size,
		}
		go SaveFileLinK(fileLink)
	}

	/*推送逻辑*/
	pushLen := 0
	var result int = OK
	var mid int64 = 0
	if strings.HasSuffix(toUserName, APP_SUFFIX) { //推应用,不走会话机制

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			logger.Error(err)
			return ParamErr, mid
		}
		logger.Infof("toKey [%v] ", toUserName)
		//TODO 推应用
		result, mid = push(toUserName, msgBytes, expire)
		if result != OK {
			return result, mid
		}
		pushLen = 1

	} else {
		isQunPush := strings.HasSuffix(toUserName, QUN_SUFFIX)
		names, _ := getNames(toUserName, sessionArgs)

		fromUserName := msg["fromUserName"].(string)
		myUserName := pushCnt.CallerId + USER_SUFFIX

		/*群消息不需要查出自己的会话，来自群消息（群通知）也不需要查出自己的会话， 自己给自己发的时候也不需查出发送者的会话*/
		/*是群推，并且自己有多个设备（多台手机）要同步群消息*/
		isFromQun := strings.HasSuffix(fromUserName, QUN_SUFFIX)
		if !isQunPush && !isFromQun && (fromUserName != toUserName) {
			myUserNames, _ := getNames(myUserName, []string{"all"})
			names = append(names, myUserNames...)
		}

		newContent := msg["content"].(string)
		originalContent := msg["content"].(string)
		originalContent = newContent[strings.Index(newContent, "&&")+2 : len(newContent)]
		// 推送
		for _, name := range names {
			msg["toUserName"] = name.Id + name.Suffix
			//发送给群同步给自己
			if isQunPush && name.Id == pushCnt.CallerId { //不能屏蔽群发送给用户的通知消息
				// 不用推送给当前设备
				deviceID, ok := msg["deviceID"].(string)
				if ok && strings.HasSuffix(name.SessionId, deviceID) {
					continue
				}

				msg["fromUserName"] = myUserName
				msg["toUserName"] = toUserName
				msg["content"] = originalContent

			} else if name.Id == pushCnt.CallerId && !isFromQun { //发送给人,同步给自己
				// 不用推送给当前设备
				deviceID, ok := msg["deviceID"].(string)
				if ok && strings.HasSuffix(name.SessionId, deviceID) {
					continue
				}
				msg["toUserName"] = toUserName
				msg["toDisplayName"] = name.DisplayName

			} else if isQunPush { // 群发时给其他用户，还原msg信息
				//复制拼接后的content
				msg["content"] = newContent
				msg["fromUserName"] = fromUserName
				msg["toDisplayName"] = name.DisplayName
			}

			key := name.toKey()

			msg["toUserKey"] = key
			msg["activeSessions"] = name.ActiveSessionIds

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				logger.Error(err)

				return ParamErr, mid
			}
			logger.Infof("toKey [%v] ", key)
			result, mid = push(key, msgBytes, expire)
			if OK != result {
				logger.Errorf("Push message failed [%v]", msg)

				// 推送分发过程中失败不立即返回，继续下一个推送
			}

		}
		pushLen = len(names)
	}
	//统计消息推送记录
	pushCnt.Count = pushLen
	go StatisticsPush(&pushCnt)

	return result, mid
}

// 按 key 推送.
func push(key string, msgBytes []byte, expire int) (ret int, mid int64) {
	node := myrpc.GetComet(key)

	if node == nil || node.CometRPC == nil {
		logger.Errorf("Get comet node failed [key=%s]", key)

		return NotFoundServer, 0
	}

	client := node.CometRPC.Get()
	if client == nil {
		logger.Errorf("Get comet node RPC client failed [key=%s]", key)

		return NotFoundServer, 0
	}

	pushArgs := &myrpc.CometPushPrivateArgs{Msg: json.RawMessage(msgBytes), Expire: uint(expire), Key: key}

	if err := client.Call(myrpc.CometServicePushPrivate, pushArgs, &mid); err != nil {
		logger.Errorf("client.Call(\"%s\", \"%v\", &ret) error(%v)", myrpc.CometServicePushPrivate, string(msgBytes), err)

		return InternalErr, 0
	}

	logger.Infof("Pushed a message to [key=%s]", key)
	return OK, mid
}

// 构造推送 name 集.
func buildNames(userIds []string, sessionArgs []string) (names []*Name) {
	for _, userId := range userIds {
		sessions := session.GetSessions(userId, sessionArgs)
		m := getUserByUid(userId)
		displayName := ""
		if nil != m {
			displayName = m.NickName
		}

		activeSessionIds := []string{}
		for _, s := range sessions {
			if session.SESSION_STATE_ACTIVE == s.State {
				activeSessionIds = append(activeSessionIds, s.Id)
			}
		}

		// id@user (i.e. for offline msg)发送离线消息使用
		name := &Name{Id: userId, SessionId: userId /* user_id 作为 session_id */, ActiveSessionIds: activeSessionIds,
			Suffix: USER_SUFFIX, DisplayName: displayName}
		names = append(names, name)

		for _, s := range sessions {
			name := &Name{Id: userId, SessionId: s.Id, ActiveSessionIds: activeSessionIds, Suffix: USER_SUFFIX, DisplayName: displayName}
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
		logger.Warn("应用推需要走单独的接口")
		return names, APP_SUFFIX
	} else {
		return names, "@UNDEFINDED"
	}
}
