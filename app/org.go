package app

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/EPICPaaS/appmsgsrv/db"
	//"github.com/EPICPaaS/appmsgsrv/session"
	"github.com/golang/glog"
)

/*成员结构体*/
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

	sql := "select id, name, nickname, status, avatar, tenant_id, name_py, name_quanpin, mobile, area from user where " + fieldName + "=?"

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
		err = row.Scan(&rec.Uid, &rec.Name, &rec.NickName, &rec.Status, &rec.Avatar, &rec.TenantId, &rec.PYInitial, &rec.PYQuanPin, &rec.Mobile, &rec.Area)
		if err != nil {
			glog.Error(err)
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
		glog.Error(err)
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

	uid := baseReq["uid"].(string)
	deviceId := baseReq["deviceID"].(string)
	deviceType := baseReq["deviceType"].(string)
	userName := args["userName"].(string)
	password := args["password"].(string)

	glog.V(5).Infof("uid [%s], deviceId [%s], deviceType [%s], userName [%s], password [%s]",
		uid, deviceId, deviceType, userName, password)

	// TODO: 登录验证逻辑
	member := getUserByCode(userName)
	if nil == member {
		baseRes.ErrMsg = "auth failed"
		baseRes.Ret = ParamErr

		return
	}

	// 登录后设置用户未关联的会话session;  sessionId:{userI_id}_{device_type}-{real_device_id}
	//sessionId := "anonymous_" + deviceType + "_" + deviceId
	//go session.UpdateSessionUserID(sessionId, member.Uid)

	//记录apnsToken
	apnsTokenStr := args["apnsToken"].(string)
	if len(apnsTokenStr) > 0 {

		apnsToken := &ApnsToken{
			UserId:    member.Uid,
			DeviceId:  deviceId,
			ApnsToken: apnsTokenStr,
			Created:   time.Now().Local(),
			Updated:   time.Now().Local(),
		}
		if !insertApnsToken(apnsToken) {
			baseRes.Ret = InternalErr
			baseRes.ErrMsg = "Sava apns_token faild"
			return
		}
	}
	// 客户端登录记录
	go Device.loginLog(&Client{UserId: member.Uid, Type: deviceType, DeviceId: deviceId})

	member.UserName = member.Uid + USER_SUFFIX

	res["uid"] = member.Uid

	token, err := genToken(member)
	if nil != err {
		glog.Error(err)

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
			glog.Error(err)
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
			glog.Error(err)
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
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)

	input := map[string]interface{}{}
	if err := json.Unmarshal(bodyBytes, &input); err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = ParamErr
		return
	}

	orgId := input["uid"].(string)
	memberList := getUserListByOrgId(orgId)
	res["memberCount"] = len(memberList)
	res["memberList"] = memberList
}

/*单位数据结构体*/
type org struct {
	id        string
	name      string
	shortName string
	parentId  string
	tenantId  string
	location  string
	sort      int
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

/*同步人员*/
func (*device) SyncUser(w http.ResponseWriter, r *http.Request) {
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
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)

	member := member{}
	if err := json.Unmarshal(bodyBytes, &member); err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = ParamErr
		return
	}

	exists := isUserExists(member.Uid)
	if exists {
		updateUser(&member, tx)
	} else {

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
func (*device) SyncOrg(w http.ResponseWriter, r *http.Request) {
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
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)

	org := org{}
	if err := json.Unmarshal(bodyBytes, &org); err != nil {
		baseRes["errMsg"] = err.Error()
		baseRes["ret"] = ParamErr
		return
	}

	exists, parentId := isExists(org.id)
	if exists && parentId == org.parentId {
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
func addOrg(org *org, tx *sql.Tx) {
	smt, err := tx.Prepare("insert into org(id, name , short_name, parent_id, tenant_id, sort) values(?,?,?,?,?,?)")
	if smt != nil {
		defer smt.Close()
	} else {
		return
	}

	if err != nil {
		return
	}

	smt.Exec(org.id, org.name, org.shortName, org.parentId, org.tenantId, org.sort)
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
		return
	}

	smt.Exec(org.name, org.shortName, org.parentId, org.sort, org.id)
}

/*设置location*/
func resetLocation2(org *org, location string, tx *sql.Tx) {
	smt, err := tx.Prepare("update org set location=? where id=?")
	if smt != nil {
		defer smt.Close()
	} else {
		return
	}

	if err != nil {
		return
	}

	smt.Exec(location, org.id)
}

/*设置location*/
func resetLocation(org *org, tx *sql.Tx) {
	if org.parentId == "" {
		resetLocation2(org, "00", tx)
	}
	smt, err := tx.Prepare("select location from org where parent_id=? order by location desc")
	if smt != nil {
		defer smt.Close()
	} else {
		return
	}

	if err != nil {
		return
	}

	row, err := smt.Query(org.parentId)
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
		resetLocation2(org, caculateLocation(loc), tx)
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

		row, _ := smt.Query(org.parentId)
		if row != nil {
			defer row.Close()
		} else {
			return
		}

		for row.Next() {
			row.Scan(&loc)
			break
		}

		resetLocation2(org, caculateLocation(loc+"$$"), tx)
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
		second = string(rs[lt-2:])
	} else {
		first = string(rs[0])
		second = string(rs[1])
	}

	if first == "$" {
		return "00"
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
		glog.Error(err)

		return false
	}

	row, err := smt.Query(fromUid, toUId)
	if nil != err {
		glog.Error(err)

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
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
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

	glog.V(1).Infof("Uid [%s], DeviceId [%s], userName [%s], password [%s]",
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

/*在用户所在租户（单位）搜索用户，根据传入的searchKey*/
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
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
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
			glog.Error(err)
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
			glog.Error(err)
		}

		rec.UserName = rec.Uid + USER_SUFFIX
		ret = append(ret, &rec)
	}

	sql = "select count(*) from user where nickname like ?"
	smt, err = db.MySQL.Prepare(sql)
	if smt != nil {
		defer smt.Close()
	} else {
		return nil, 0
	}

	if err != nil {
		return nil, 0
	}

	row, err = smt.Query("%" + nickName + "%")
	if row != nil {
		defer row.Close()
	} else {
		return nil, 0
	}

	cnt := 0
	for row.Next() {
		err = row.Scan(&cnt)
		if err != nil {
			glog.Error(err)
		}
	}
	return ret, cnt
}
