package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"text/template"
	"time"
)

// 成员结构体.
type member struct {
	Uid         string    `json:"uid"`
	UserName    string    `json:"userName"`
	Name        string    `json:"name"`
	NickName    string    `json:"nickName"`
	HeadImgUrl  string    `json:"headImgUrl"`
	MemberCount int       `json:"memberCount"`
	MemberList  []*member `json:"memberList"`
	PYInitial   string    `json:"pYInitial"`
	PYQuanPin   string    `json:"pYQuanPin"`
	Status      string    `json:"status"`
	StarFriend  int       `json:"starFriend"`
	Avatar      string    `json:"avatar"`
	parentId    string    `json:"parentId"`
	Sort        int       `json:"sort"`
	rand        int       `json:"rand"`
	Password    string    `json:"password"`
	TenantId    string    `json:"tenantId"`
	Email       string    `json:"email"`
	Mobile      string    `json:"mobile"`
	Area        string    `json:"area"`
}
type Tenant struct {
	Id         string    `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Status     int       `json:"status"`
	CustomerId string    `json:"customerId"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
}

type ExternalInterface struct {
	Id         string    `json:"id"`
	CustomerId string    `json:"customerId"`
	Type       string    `json:"type"`
	Owner      int       `json:"owner"`
	HttpUrl    string    `json:"httpUrl"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
}

// 用户身份验证接口.
//
//  1. 根据指定的 tenantId 查询 customerId
//  2. 在 external_interface 表中根据 customerId、type = 'login' 等信息查询接口地址
//  3. 根据接口地址调用验证接口
func loginAuth(username, password, customer_id string) (loginOk bool, user *member) {

	// TODO: 旭东
	EI := GetExtInterface(customer_id, "login")
	if EI != nil {
		if EI.Owner == 0 { //自己的登录
			user = getUserByCode(username)
			if user != nil && user.Password == password {
				return true, user
			} else {
				return false, nil
			}
		} else {
			data := []byte(`{
					     "userName":` + username + `,
					     "password":` + password +
				`}`)
			body := bytes.NewReader(data)
			res, err := http.Post(EI.HttpUrl, "text/plain;charset=UTF-8", body)
			if err != nil {
				logger.Error(err)
				return false, nil
			}

			resBodyByte, err := ioutil.ReadAll(res.Body)
			defer res.Body.Close()
			if err != nil {
				logger.Error(err)
				return false, nil
			}
			var respBody map[string]interface{}

			if err := json.Unmarshal(resBodyByte, &respBody); err != nil {
				logger.Errorf("convert to json failed (%s)", err.Error())
				return false, nil
			}
			success, ok := respBody["success"].(bool)
			if ok && success {
				/*获取用户，同步*/
				userBody, err := json.Marshal(respBody["user"])
				if err != nil {
					logger.Errorf("convert to json failed (%s)", err.Error())
					return false, nil
				}

				if err := json.Unmarshal(userBody, &user); err != nil {
					logger.Errorf("convert to json failed (%s)", err.Error())
					return false, nil
				}

				exists := isUserExists(user.Uid)
				if exists {
					//有则更新
					if !updateMember(user) {
						return false, nil
					}
				} else {
					//新增
					if !addUser(user) {
						return false, nil
					}
				}

				/*获取租户信息，同步*/
				tenantBody, err := json.Marshal(respBody["tenant"])
				if err != nil {
					logger.Errorf("convert to json failed (%s)", err.Error())
					return false, nil
				}
				var tenant Tenant
				if err := json.Unmarshal(tenantBody, &tenant); err != nil {
					logger.Errorf("convert to json failed (%s)", err.Error())
					return false, nil
				}
				tenant.CustomerId = customer_id
				if !saveTennat(&tenant) {
					logger.Error("登录设置tenant失败！")
					return false, nil
				}

				//登录成功
				return true, user
			} else {
				return false, nil
			}
		}
	}
	return false, nil
}

/*根据userId获取成员信息*/
func getUserByUid(uid string) *member {
	return getUserByField("id", uid)
}

/*根据 email或者name 获取成员信息,传入的code带@符号时是为email*/
func getUserByCode(code string) *member {
	isEmail := false
	if strings.LastIndex(code, "@") > -1 {
		isEmail = true
	}
	fieldName := "name"
	if isEmail {
		fieldName = "email"
	}
	return getUserByField(fieldName, code)
}

/*根据传入的筛选列fieldName和参数fieldArg查询成员*/
func getUserByField(fieldName, fieldArg string) *member {

	sql := "select id, name, nickname, status, avatar, tenant_id, name_py, name_quanpin,password, mobile, area from user where " + fieldName + "=?"

	smt, err := db.MySQL.Prepare(sql)
	if smt != nil {
		defer smt.Close()
	} else {
		return nil
	}

	if err != nil {
		return nil
	}

	row, err := smt.Query(fieldArg)
	if row != nil {
		defer row.Close()
	} else {
		return nil
	}

	for row.Next() {
		rec := member{}
		err = row.Scan(&rec.Uid, &rec.Name, &rec.NickName, &rec.Status, &rec.Avatar, &rec.TenantId, &rec.PYInitial, &rec.PYQuanPin, &rec.Password, &rec.Mobile, &rec.Area)
		if err != nil {
			logger.Error(err)
		}

		rec.UserName = rec.Uid + USER_SUFFIX
		return &rec
	}

	return nil
}

// 用户二维码处理，返回用户信息 HTML.
func UserErWeiMa(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	uid := ""
	if len(r.Form) > 0 {
		uid = r.Form.Get("id")
	}

	user := getUserByUid(uid)
	if nil == user {
		http.Error(w, "Not Found", 404)

		return
	}

	t, err := template.ParseFiles("view/erweima.html")

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	model := map[string]interface{}{
		"staticServer": "/app/static",
		"nickname":     user.NickName, "username": user.NickName, "email": user.Email, "phone": user.Mobile}

	t.Execute(w, model)
}

// 根据 UserName 获取用户信息.
func (*device) GetMemberByUserName(w http.ResponseWriter, r *http.Request) {
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
	user := getUserByToken(token)
	if nil == user {
		baseRes.Ret = AuthErr

		return
	}

	userName := args["userName"].(string)
	uid := userName[:strings.LastIndex(userName, USER_SUFFIX)]

	toUser := getUserByUid(uid)
	if nil == toUser {
		baseRes.Ret = NotFound

		return
	}

	// 是否是常用联系人
	if isStar(user.Uid, toUser.Uid) {
		toUser.StarFriend = 1
	}

	res["member"] = toUser
}

// 客户端设备登录.
func (*device) Login(w http.ResponseWriter, r *http.Request) {
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

	uid := baseReq["uid"].(string)
	deviceId := baseReq["deviceID"].(string)
	customer_id := baseReq["customer_id"].(string)
	deviceType := baseReq["deviceType"].(string)
	userName := args["userName"].(string)
	password := args["password"].(string)

	logger.Tracef("uid [%s], deviceId [%s], deviceType [%s], userName [%s], password [%s]",
		uid, deviceId, deviceType, userName, password)

	loginOK, member := loginAuth(userName, password, customer_id)
	if !loginOK {
		baseRes.ErrMsg = "auth failed"
		baseRes.Ret = LoginErr
		return
	}

	//记录apnsToken
	apnsTokenStr, ok := args["apnsToken"].(string)

	if ok {
		apnsToken := &ApnsToken{
			UserId: member.Uid,
			//DeviceId:  deviceId,  IOS获取不到deviceID，用APNSToken代替
			DeviceId:  apnsTokenStr,
			ApnsToken: apnsTokenStr,
			Created:   time.Now().Local(),
			Updated:   time.Now().Local(),
		}

		// 先删除该deviceId
		deleteApnsTokenByDeviceId(apnsToken.DeviceId)
		//再插入该设备对应的用户
		if !insertApnsToken(apnsToken) {
			baseRes.Ret = InternalErr
			baseRes.ErrMsg = "Save apns_token faild"
			return
		}
	}
	// 客户端登录记录
	go Device.loginLog(&Client{UserId: member.Uid, Type: deviceType, DeviceId: deviceId})

	member.UserName = member.Uid + USER_SUFFIX

	res["uid"] = member.Uid

	token, err := genToken(member)
	if nil != err {
		logger.Error(err)

		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr

		return
	}

	res["token"] = token
	res["member"] = member
}

//网页客户端登录
func (*appWeb) WebLogin(w http.ResponseWriter, r *http.Request) {

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

		logger.Info(w, *callback, "(", resJsonStr, ")")
	}()

	//获取请求数据
	r.ParseForm()
	tmp := r.FormValue("callbackparam")
	callback = &tmp
	uid := r.FormValue("baseRequest[uid]")
	deviceType := r.FormValue("baseRequest[deviceType]")
	userName := r.FormValue("userName")
	password := r.FormValue("password")

	logger.Tracef("uid [%s], deviceType [%s], userName [%s], password [%s]",
		uid, deviceType, userName, password)

	member := getUserByUid(uid)
	if nil == member || member.Password != password {
		baseRes.ErrMsg = "auth failed"
		baseRes.Ret = ParamErr
		return
	}

	member.UserName = member.Uid + USER_SUFFIX
	res["uid"] = member.Uid
	token, err := genToken(member)
	if nil != err {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr
		return
	}

	res["token"] = token
	res["member"] = member
}

type members []*member

type BySort struct {
	memberList members
}

/*获取成员总数*/
func (s BySort) Len() int { return len(s.memberList) }

/*交换成员顺序*/
func (s BySort) Swap(i, j int) {
	s.memberList[i], s.memberList[j] = s.memberList[j], s.memberList[i]
}

/*判断两个成员的顺序*/
func (s BySort) Less(i, j int) bool {
	return s.memberList[i].Sort < s.memberList[j].Sort
}

func sortMemberList(lst []*member) {
	sort.Sort(BySort{lst})

	for _, rec := range lst {
		sort.Sort(BySort{rec.MemberList})
	}
}

/*根据租户id（TenantId）获取成员*/
func getUserListByTenantId(id string) members {
	smt, err := db.MySQL.Prepare("select id, name, nickname, status, avatar, tenant_id, name_py, name_quanpin, mobile, area from user where tenant_id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		return nil
	}

	if err != nil {
		return nil
	}

	row, err := smt.Query(id)
	if row != nil {
		defer row.Close()
	} else {
		return nil
	}
	ret := members{}
	for row.Next() {
		rec := new(member)
		err = row.Scan(&rec.Uid, &rec.Name, &rec.NickName, &rec.Status, &rec.Avatar, &rec.TenantId, &rec.PYInitial, &rec.PYQuanPin, &rec.Mobile, &rec.Area)
		if err != nil {
			logger.Error(err)
		}
		ret = append(ret, rec)
	}

	return ret
}

/*根据单位id（TenantId）获取成员*/
func getUserListByOrgId(id string) members {
	smt, err := db.MySQL.Prepare("select `user`.`id`, `user`.`name`, `user`.`nickname`, `user`.`status`, `user`.`avatar`, `user`.`tenant_id`, `user`.`name_py`, `user`.`name_quanpin`, `user`.`mobile`, `user`.`area`,`org_user`.`sort`	from `user`,`org_user` where `user`.`id`=`org_user`.`user_id` and org_id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		return nil
	}

	if err != nil {
		return nil
	}

	row, err := smt.Query(id)
	if row != nil {
		defer row.Close()
	} else {
		return nil
	}
	ret := members{}
	for row.Next() {
		rec := new(member)
		err = row.Scan(&rec.Uid, &rec.Name, &rec.NickName, &rec.Status, &rec.Avatar, &rec.TenantId, &rec.PYInitial, &rec.PYQuanPin, &rec.Mobile, &rec.Area, &rec.Sort)
		if err != nil {
			logger.Error(err)
		}
		rec.UserName = rec.Uid + USER_SUFFIX
		ret = append(ret, rec)
	}
	return ret
}

/*获取单位的人员信息*/
func (*device) GetOrgUserList(w http.ResponseWriter, r *http.Request) {
	baseRes := map[string]interface{}{"ret": OK, "errMsg": ""}

	body := ""
	res := map[string]interface{}{"baseResponse": baseRes}
	defer RetPWriteJSON(w, r, res, &body, time.Now())

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = ParamErr
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)

	input := map[string]interface{}{}
	if err := json.Unmarshal(bodyBytes, &input); err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = ParamErr
		return
	}

	orgId := input["orgid"].(string)
	memberList := getUserListByOrgId(orgId)
	res["memberCount"] = len(memberList)
	res["memberList"] = memberList
}

/*获取单位的人员信息*/
func (*app) GetOrgUserList(w http.ResponseWriter, r *http.Request) {
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
		logger.Errorf("Application[%v]  AuthErr  [%v]", application.Name, err)
		return
	}

	orgId := args["orgid"].(string)
	memberList := getUserListByOrgId(orgId)
	res["memberCount"] = len(memberList)
	res["memberList"] = memberList
}

/*获取人员的单位集*/
func (*app) GetOrgList(w http.ResponseWriter, r *http.Request) {
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
		logger.Errorf("Application[%v]  AuthErr  [%v]", application.Name, err)
		return
	}

	userId := args["uid"].(string)
	orgList := getOrgListByUserId(userId)
	res["orgCount"] = len(orgList)
	res["orgList"] = orgList
}

func getOrgListByUserId(userId string) []*org {
	rows, _ := db.MySQL.Query("select * from `org`  where  `org`.`id`  in  (select `org_user`.`org_id`  from `org_user` where `org_user`.`user_id` = ?) ", userId)
	if rows != nil {
		defer rows.Close()
	}
	ret := []*org{}
	for rows.Next() {
		resource := &org{}
		if err := rows.Scan(&resource.ID, &resource.Name, &resource.ShortName, &resource.ParentId, &resource.Location, &resource.TenantId, &resource.Sort); err != nil {
			logger.Error(err)

			return nil
		}
		ret = append(ret, resource)
	}
	return ret
}

/*添加用户-组织关联关系*/
func (*app) AddOrgUser(w http.ResponseWriter, r *http.Request) {
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
		logger.Errorf("Application[%v]  AuthErr  [%v]", application.Name, err)
		return
	}

	orgId := args["orgid"].(string)
	userId := args["uid"].(string)
	b := addOrgUser(orgId, userId)
	res["successed"] = b
}

/*移除用户-组织关联关系*/
func (*app) RemoveOrgUser(w http.ResponseWriter, r *http.Request) {
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
		logger.Errorf("Application[%v]  AuthErr  [%v]", application.Name, err)
		return
	}

	orgId := args["orgid"].(string)
	userId := args["uid"].(string)
	b := removeOrgUser(orgId, userId)
	res["successed"] = b
}

func removeOrgUser(orgId, userId string) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		logger.Error(err)

		return false
	}

	_, err = tx.Exec("delete from org_user where org_id = ? and user_id = ?", orgId, userId)
	if err != nil {
		logger.Error(err)

		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}

		return false
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err)

		return false
	}

	return true
}

/*单位数据结构体*/
type org struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	ParentId  string `json:"parentId"`
	TenantId  string `json:"tenantId"`
	Location  string `json:"location"`
	Sort      int    `json:"sort"`
}

/*修改用户信息*/
func updateUser(member *member, tx *sql.Tx) error {
	st, err := tx.Prepare("update user set name=?, nickname=?, avastar=?, name_py=?, name_quanpin=?, status=?, rand=?, password=?, tenant_id=?, updated=?, email=? where id=?")
	if err != nil {
		return err
	}

	_, err = st.Exec(member.Name, member.NickName, member.Avatar, member.PYInitial, member.PYQuanPin, member.Status, member.rand, member.Password, member.TenantId, time.Now(), member.Email, member.Uid)

	return err
}

/*修改用户信息*/
func updateMember(member *member) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}

	_, err = tx.Exec("update user set name=?, nickname=?, avatar=?, name_py=?, name_quanpin=?, status=?, rand=?, password=?, tenant_id=?, updated=?, email=? where id=?", member.Name, member.NickName, member.Avatar, member.PYInitial, member.PYQuanPin, member.Status, member.rand, member.Password, member.TenantId, time.Now(), member.Email, member.Uid)
	if err != nil {

		logger.Error(err)
		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return false
	}
	//提交操作
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return false
	}
	return true
}

/*添加用户信息*/
func addUser(member *member) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}
	_, err = tx.Exec("insert into user(id,name,nickname,avatar,name_py,name_quanpin,status,password,tenant_id,email,mobile,area,created,updated)values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", member.Uid, member.Name, member.NickName, member.Avatar, member.PYInitial, member.PYQuanPin, member.Status, member.Password, member.TenantId, member.Email, member.Mobile, member.Area, time.Now(), time.Now())
	if err != nil {
		logger.Error(err)

		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return false
	}
	//提交操作
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return false
	}
	return true
}

//保存人员与单位关系
func addOrgUser(orgId, userId string) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}

	_, err = tx.Exec("insert into org_user(id,org_id,user_id) values(?,?,?)", uuid.New(), orgId, userId)
	if err != nil {
		logger.Error(err)
		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return false
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return false
	}
	return true
}

/*同步人员*/
func (*app) SyncUser(w http.ResponseWriter, r *http.Request) {

	baseRes := map[string]interface{}{"ret": OK, "errMsg": ""}
	tx, err := db.MySQL.Begin()

	body := ""
	res := map[string]interface{}{"baseResponse": baseRes}
	defer RetPWriteJSON(w, r, res, &body, time.Now())

	if err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = ParamErr
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}

	body = string(bodyBytes)

	var args map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &args); err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = ParamErr
		return
	}

	//应用校验
	baseReq := args["baseRequest"].(map[string]interface{})
	token := baseReq["token"].(string)
	_, err = getApplicationByToken(token)
	if nil != err {
		baseRes["ret"] = AuthErr
		baseRes["errMsg"] = "Authorization failure"
		return
	}

	orgId := args["orgId"].(string)
	memberMap := args["member"].(map[string]interface{})

	menberObj := &member{
		Uid:       memberMap["uid"].(string),
		Name:      memberMap["name"].(string),
		NickName:  memberMap["nickName"].(string),
		PYInitial: memberMap["pYInitial"].(string),
		PYQuanPin: memberMap["pYQuanPin"].(string),
		Status:    memberMap["status"].(string),
		Avatar:    memberMap["avatar"].(string),
		Password:  memberMap["password"].(string),
		TenantId:  memberMap["tenantId"].(string),
		Email:     memberMap["email"].(string),
		Mobile:    memberMap["mobile"].(string),
		Area:      memberMap["area"].(string),
	}

	exists := isUserExists(menberObj.Uid)
	if exists {
		//有则更新
		updateUser(menberObj, tx)
	} else {
		//新增
		resFlag := addUser(menberObj)
		//添加单位人员关系
		if len(orgId) > 0 {
			resFlag = addOrgUser(orgId, menberObj.Uid)
		}
		if !resFlag {
			baseRes["ret"] = InternalErr
			baseRes["errMsg"] = "sysnUser  failure"
		}
	}

	rerr := recover()
	if rerr != nil {
		baseRes["errMsg"] = rerr
		baseRes["ret"] = InternalErr
		tx.Rollback()
	} else {
		err = tx.Commit()
		if err != nil {
			baseRes["errMsg"] = err.Error()
			baseRes["ret"] = InternalErr
		}
	}
}

/*同步单位*/
func (*app) SyncOrg(w http.ResponseWriter, r *http.Request) {
	baseRes := map[string]interface{}{"ret": OK, "errMsg": ""}
	tx, err := db.MySQL.Begin()

	body := ""
	res := map[string]interface{}{"baseResponse": baseRes}
	defer RetPWriteJSON(w, r, res, &body, time.Now())

	if err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		logger.Error(err)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = ParamErr
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}

	body = string(bodyBytes)

	var args map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &args); err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = ParamErr
		logger.Error(err)
		return
	}

	//应用校验
	baseReq := args["baseRequest"].(map[string]interface{})
	token := baseReq["token"].(string)
	_, err = getApplicationByToken(token)
	if nil != err {
		baseRes["ret"] = AuthErr
		baseRes["errMsg"] = "authfailure"
		logger.Error(err)
		return
	}

	orgMap := args["org"].(map[string]interface{})
	org := org{
		ID:        orgMap["id"].(string),
		Name:      orgMap["name"].(string),
		ShortName: orgMap["shortName"].(string),
		ParentId:  orgMap["parentId"].(string),
		TenantId:  orgMap["tenantId"].(string),
		Sort:      int(orgMap["sort"].(float64)),
	}
	exists, parentId := isExists(org.ID)
	if exists && parentId == org.ParentId {
		updateOrg(&org, tx)
	} else if exists {
		updateOrg(&org, tx)
		resetLocation(&org, tx)
	} else {
		addOrg(&org, tx)
		resetLocation(&org, tx)
	}

	rerr := recover()
	if rerr != nil {
		baseRes["errMsg"] = rerr
		baseRes["ret"] = InternalErr
		tx.Rollback()
	} else {
		err = tx.Commit()
		if err != nil {
			baseRes["errMsg"] = err.Error()
			baseRes["ret"] = InternalErr
		}
	}
}

/*新增单位*/
func addOrg(org *org, tx *sql.Tx) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}
	_, err = tx.Exec("insert into org(id, name , short_name, parent_id, tenant_id, sort) values(?,?,?,?,?,?)", org.ID, org.Name, org.ShortName, org.ParentId, org.TenantId, org.Sort)
	if err != nil {
		logger.Error(err)
		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return false
	}
	if err = tx.Commit(); err != nil {
		logger.Error(err)
		return false
	}
	return true
}

/*修改单位信息*/
func updateOrg(org *org, tx *sql.Tx) {
	smt, err := tx.Prepare("update org set name=?, short_name=?, parent_id=?, sort=? where id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		return
	}

	if err != nil {
		logger.Error(err)
		return
	}

	smt.Exec(org.Name, org.ShortName, org.ParentId, org.Sort, org.ID)

}

/*设置location*/
func resetLocation2(org *org, location string) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}

	_, err = tx.Exec("update org set location=? where id=?", location, org.ID)
	if err != nil {
		logger.Error(err)
		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return false
	}

	if err = tx.Commit(); err != nil {
		logger.Error(err)
		return false
	}
	return true
}

/*设置location*/
func resetLocation(org *org, tx *sql.Tx) {
	if org.ParentId == "" {
		resetLocation2(org, "00")
	}
	smt, err := tx.Prepare("select location  from org where parent_id=? and id !=? order by location desc")
	if smt != nil {
		defer smt.Close()
	} else {
		return
	}

	if err != nil {
		return
	}

	row, err := smt.Query(org.ParentId, org.ID)
	if row != nil {
		defer row.Close()
	} else {
		return
	}

	// FIXME: 李旭东
	loc := ""
	hasBrother := false
	for row.Next() {
		row.Scan(&loc)
		hasBrother = true
		break
	}
	/*如果有兄弟部门，则通过上一个兄弟部门location（用于本地树结构关系）计算出自己的location ; 没有则通过父亲的计算location*/
	if hasBrother {
		resetLocation2(org, caculateLocation(loc))
	} else {

		smt, err = tx.Prepare("select location from org where id=?")
		if smt != nil {
			defer smt.Close()
		} else {
			return
		}

		if err != nil {
			return
		}

		row, _ := smt.Query(org.ParentId)
		if row != nil {
			defer row.Close()
		} else {
			return
		}

		for row.Next() {
			row.Scan(&loc)
			break
		}

		resetLocation2(org, caculateLocation(loc+"$$"))
	}
}

/*计算出location，用于树的层级关系*/
func caculateLocation(loc string) string {
	rs := []rune(loc)
	lt := len(rs)
	prefix := ""
	first := ""
	second := ""
	if lt > 2 {
		prefix = string(rs[:(lt - 2)])
		first = string(rs[(lt - 2):(lt - 1)])
		second = string(rs[lt-1:])
	} else {
		first = string(rs[0])
		second = string(rs[1])
	}

	// FIXME: 李旭东
	if first == "$" { // 通过父亲生成location
		return prefix + "00"
	} else {
		return prefix + nextLocation(first, second)
	}
}

/*递增出下一个同级location*/
func nextLocation(first, second string) string {
	if second == "9" {
		second = "a"
	} else {
		if second == "z" {
			second = "0"
			if first == "9" {
				first = "a"
			} else {
				bf := first[0]
				bf++
				first = string(bf)
			}
		} else {
			bs := second[0]
			bs++
			second = string(bs)
		}
	}
	return first + second
}

/*通过userId判断该用户是否存在*/
func isUserExists(id string) bool {
	smt, err := db.MySQL.Prepare("select 1 from user where id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		return false
	}

	if err != nil {
		return false
	}

	row, err := smt.Query(id)
	if row != nil {
		defer row.Close()
	} else {
		return false
	}

	for row.Next() {
		return true
	}

	return false
}

/*判断两个用户是否为常联系人*/
func isStar(fromUid, toUId string) bool {
	smt, err := db.MySQL.Prepare("select 1 from user_user where from_user_id=? and to_user_id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		return false
	}

	if err != nil {
		logger.Error(err)

		return false
	}

	row, err := smt.Query(fromUid, toUId)
	if nil != err {
		logger.Error(err)

		return false
	}

	if row != nil {
		defer row.Close()
	} else {
		return false
	}

	return row.Next()
}

/*判断单位是否存在，且返回他的父节点id*/
func isExists(id string) (bool, string) {
	smt, err := db.MySQL.Prepare("select parent_id from org where id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		return false, ""
	}

	if err != nil {
		return false, ""
	}

	row, err := smt.Query(id)
	if row != nil {
		defer row.Close()
	} else {
		return false, ""
	}

	for row.Next() {
		parentId := ""
		row.Scan(&parentId)
		return true, parentId
	}

	return false, ""
}

//获取当前用户的单位信息（完整的单位部门树）和用户好友
func (*device) GetOrgInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	baseRes := map[string]interface{}{"ret": OK, "errMsg": ""}
	body := ""
	res := map[string]interface{}{"baseResponse": baseRes}
	defer RetPWriteJSON(w, r, res, &body, time.Now())

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = ParamErr
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)

	var args map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &args); err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = ParamErr
		return
	}

	baseReq := args["baseRequest"].(map[string]interface{})

	uid := baseReq["uid"].(string)
	deviceId := baseReq["deviceID"]
	userName := args["userName"]
	password := args["password"]

	// Token 校验
	token := baseReq["token"].(string)
	currentUser := getUserByToken(token)
	if nil == currentUser {
		baseRes["ret"] = AuthErr
		baseRes["errMsg"] = "会话超时请重新登录"
		return
	}

	logger.Tracef("Uid [%s], DeviceId [%s], userName [%s], password [%s]",
		uid, deviceId, userName, password)

	smt, err := db.MySQL.Prepare("select id, name,  parent_id, sort from org where tenant_id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		return
	}

	if err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		return
	}

	//查询出该租户（组织）的所有部门
	row, err := smt.Query(currentUser.TenantId)
	if row != nil {
		defer row.Close()
	} else {
		return
	}
	data := members{}
	for row.Next() {
		rec := new(member)
		row.Scan(&rec.Uid, &rec.NickName, &rec.parentId, &rec.Sort)
		rec.Uid = rec.Uid
		rec.UserName = rec.Uid + ORG_SUFFIX
		data = append(data, rec)
	}
	err = row.Err()
	if err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		return
	}

	unitMap := map[string]*member{}
	for _, ele := range data {
		unitMap[ele.Uid] = ele
	}

	//构造部门树结构
	rootList := []*member{}
	for _, val := range unitMap {
		if val.parentId == "" {
			rootList = append(rootList, val)
		} else {
			parent := unitMap[val.parentId]
			if parent == nil {
				continue
			}
			parent.MemberList = append(parent.MemberList, val)
			parent.MemberCount++
		}
	}

	tenant := new(member)
	res["ognizationMemberList"] = tenant
	sortMemberList(rootList)
	tenant.MemberList = rootList
	tenant.MemberCount = len(rootList)
	/*获取用户所属租户（单位）信息*/
	smt, err = db.MySQL.Prepare("select id, code, name from tenant where id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		return
	}

	if err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		return
	}

	row, err = smt.Query(currentUser.TenantId)
	if row != nil {
		defer row.Close()
	} else {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		return
	}

	for row.Next() {
		row.Scan(&tenant.Uid, &tenant.UserName, &tenant.NickName)
		tenant.UserName = tenant.Uid + TENANT_SUFFIX
		break
	}
	/*查出用户所属部门*/
	smt, err = db.MySQL.Prepare("select org_id from org_user where user_id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		return
	}

	if err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		return
	}

	row, err = smt.Query(currentUser.Uid)
	if row != nil {
		defer row.Close()
	} else {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = InternalErr
		return
	}

	for row.Next() {
		userOgnization := ""
		row.Scan(&userOgnization)
		res["userOgnization"] = userOgnization
		break
	}

	starMemberList := getStarUser(currentUser.Uid)
	res["starMemberCount"] = len(starMemberList)
	res["starMemberList"] = starMemberList
}

/*在用户所在租户（单位）搜索用户，根据传入的@searchKey，并支持分页（@offset，@limit）*/
func (*device) SearchUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	baseRes := map[string]interface{}{"ret": OK, "errMsg": ""}
	body := ""
	res := map[string]interface{}{"baseResponse": baseRes}
	defer RetPWriteJSON(w, r, res, &body, time.Now())

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res["ret"] = ParamErr
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)

	var args map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &args); err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = ParamErr
		return
	}

	baseReq := args["baseRequest"].(map[string]interface{})

	//uid := baseReq["uid"].(string)
	//deviceId := baseReq["deviceID"]
	//userName := args["userName"]
	//password := args["password"]

	// Token 校验
	token := baseReq["token"].(string)
	currentUser := getUserByToken(token)
	if nil == currentUser {
		baseRes["ret"] = AuthErr
		return
	}

	searchKey := args["searchKey"]
	searchType := args["searchType"]
	offset := args["offset"]
	limit := args["limit"]

	var memberList members
	var cnt int
	switch searchType {
	case "user":
		memberList, cnt = searchUser(currentUser.TenantId, searchKey.(string), int(offset.(float64)), int(limit.(float64)))
	case "app":
		break
	}

	res["memberListSize"] = len(memberList)
	res["memberList"] = memberList
	res["count"] = cnt
	return
}

/*获取用户所有好友信息*/
func getStarUser(userId string) members {
	ret := members{}
	sql := "select id, name, nickname, status, avatar, tenant_id, name_py, name_quanpin, mobile, area from user where id in (select to_user_id from user_user where from_user_id=?)"

	smt, err := db.MySQL.Prepare(sql)
	if smt != nil {
		defer smt.Close()
	} else {
		return nil
	}

	if err != nil {
		return nil
	}

	row, err := smt.Query(userId)
	if row != nil {
		defer row.Close()
	} else {
		return nil
	}

	for row.Next() {
		rec := member{}
		err = row.Scan(&rec.Uid, &rec.Name, &rec.NickName, &rec.Status, &rec.Avatar, &rec.TenantId, &rec.PYInitial, &rec.PYQuanPin, &rec.Mobile, &rec.Area)
		if err != nil {
			logger.Error(err)
		}

		rec.UserName = rec.Uid + USER_SUFFIX
		ret = append(ret, &rec)
	}

	return ret
}

/*通过name搜索用户，返回搜索结果（带分页），和结果条数*/
func searchUser(tenantId, nickName string, offset, limit int) (members, int) {
	ret := members{}
	sql := "select id, name, nickname, status, avatar, tenant_id, name_py, name_quanpin, mobile, area from user where tenant_id=? and nickname like ? limit ?, ?"

	smt, err := db.MySQL.Prepare(sql)
	if smt != nil {
		defer smt.Close()
	} else {
		return nil, 0
	}

	if err != nil {
		return nil, 0
	}

	row, err := smt.Query(tenantId, "%"+nickName+"%", offset, limit)
	if row != nil {
		defer row.Close()
	} else {
		return nil, 0
	}

	for row.Next() {
		rec := member{}
		err = row.Scan(&rec.Uid, &rec.Name, &rec.NickName, &rec.Status, &rec.Avatar, &rec.TenantId, &rec.PYInitial, &rec.PYQuanPin, &rec.Mobile, &rec.Area)
		if err != nil {
			logger.Error(err)
		}

		rec.UserName = rec.Uid + USER_SUFFIX
		ret = append(ret, &rec)
	}

	sql = "select count(*) from user where tenant_id=?  and  nickname like ?"
	smt, err = db.MySQL.Prepare(sql)
	if smt != nil {
		defer smt.Close()
	} else {
		return nil, 0
	}

	if err != nil {
		return nil, 0
	}

	row, err = smt.Query(tenantId, "%"+nickName+"%")
	if row != nil {
		defer row.Close()
	} else {
		return nil, 0
	}

	cnt := 0
	for row.Next() {
		err = row.Scan(&cnt)
		if err != nil {
			logger.Error(err)
		}
	}
	return ret, cnt
}

/*同步租户*/
func (*app) SyncTenant(w http.ResponseWriter, r *http.Request) {

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
	token := baseReq["token"].(string)

	//应用校验
	_, err = getApplicationByToken(token)
	if nil != err {
		baseRes.Ret = AuthErr
		baseRes.ErrMsg = "auth failure"
		return
	}
	tennatMap := args["tennat"].(map[string]interface{})
	if len(tennatMap) == 0 {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = ParamErr
		return
	}

	tenant := &Tenant{
		Id:         tennatMap["id"].(string),
		Code:       tennatMap["code"].(string),
		Name:       tennatMap["name"].(string),
		CustomerId: tennatMap["customerId"].(string),
		Status:     int(tennatMap["status"].(float64)),
	}

	if !saveTennat(tenant) {
		baseRes.Ret = InternalErr
		baseRes.ErrMsg = "synchronize tennat failure "
		return
	}
	return
}

/*添加租户*/
func saveTennat(tenant *Tenant) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}

	//修改
	if isExistTennat(tenant.Id) {
		_, err = tx.Exec("update tenant set code = ?,name=?,status=?,customer_id=?,updated=? where id =?", tenant.Code, tenant.Name, tenant.Status, tenant.CustomerId, time.Now().Local(), tenant.Id)
	} else {
		_, err = tx.Exec("insert into tenant(id,code,name,status,customer_id,created,updated) values(?,?,?,?,?,?,?)", tenant.Id, tenant.Code, tenant.Name, tenant.Status, tenant.CustomerId, time.Now().Local(), time.Now().Local())

	}

	if err != nil {
		logger.Error(err)
		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return false
	}

	if err = tx.Commit(); err != nil {
		logger.Error(err)
		return false
	}

	return true
}

//判断租户是否存在
func isExistTennat(id string) bool {
	rows, err := db.MySQL.Query("select * from tenant where id =?", id)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		logger.Error(err)
		return false
	}
	return rows.Next()
}

//更具id获取租户信息
func getTenantById(id string) *Tenant {

	rows, err := db.MySQL.Query("select id , code , name,status,customer_id,created,updated from tenant where id =?", id)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		logger.Error(err)
		return nil
	}
	for rows.Next() {
		tenant := &Tenant{}
		if err := rows.Scan(&tenant.Id, &tenant.Code, &tenant.Name, &tenant.Status, &tenant.CustomerId, &tenant.Created, &tenant.Updated); err != nil {
			logger.Error(err)
			return nil
		}

		return tenant
	}
	return nil
}

/*同步租户*/
func (*app) UserAuth(w http.ResponseWriter, r *http.Request) {

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
	appToken := baseReq["token"].(string)
	//应用校验
	_, err = getApplicationByToken(appToken)
	if nil != err {
		baseRes.Ret = AuthErr
		baseRes.ErrMsg = "auth failure"
		return
	}

	token := args["token"].(string)
	uid := args["uid"].(string)
	//用户校验
	mem := getUserByToken(token)

	if nil == mem {
		baseRes.Ret = AuthErr
		baseRes.ErrMsg = "auth failure"
		return
	}

	if uid != mem.Uid {
		baseRes.Ret = AuthErr
		baseRes.ErrMsg = "auth failure"
		return
	}
}

//根据customer_id 和type获取客户 提供的结构地址
func GetExtInterface(customer_id, Type string) *ExternalInterface {

	rows, err := db.MySQL.Query("select id , customer_id , type ,owner,http_url,created,updated from external_interface where customer_id = ? and type =?", customer_id, Type)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		logger.Error(err)
		return nil
	}
	for rows.Next() {
		ei := &ExternalInterface{}
		if err := rows.Scan(&ei.Id, &ei.CustomerId, &ei.Type, &ei.Owner, &ei.HttpUrl, &ei.Created, &ei.Updated); err != nil {
			logger.Error(err)
			return nil
		}

		return ei
	}
	return nil
}
