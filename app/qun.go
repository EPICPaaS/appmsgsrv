package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
)

const (
	// 群插入 SQL.
	InsertQunSQL = "INSERT INTO `qun` (`id`, `creator_id`, `name`, `description`, `max_member`, `avatar`, `created`, `updated`) VALUES " +
		"(?, ?, ?, ?, ?, ?, ?, ?)"
	// 群-用户关联插入 SQL.
	InsertQunUserSQL = "INSERT INTO `qun_user` (`id`, `qun_id`, `user_id`, `sort`, `role`, `created`, `updated`) SELECT ?, ?, ?, ?, ?, ?, ? FROM DUAL WHERE NOT EXISTS (SELECT 1 from `qun_user` WHERE `qun_id`= ? AND `user_id`= ?)"

	// 根据群 id 查询群内用户.
	SelectQunUserSQL = "SELECT `id`, `name`, `nickname`, `status`, `avatar`, `tenant_id`, `name_py`, `name_quanpin`, `mobile`, `area` FROM `user` where `id` in (SELECT `user_id` FROM `qun_user` where `qun_id` = ?)"
	// 根据群 id 查询群内用户 id.
	SelectQunUserIdSQL = "SELECT `user_id` FROM `qun_user` where `qun_id` = ?"
	// 根据群 id 获取群
	SelectQunById = "SELECT * FROM `qun` where `id` = ?"
	//根据群id修改群topic
	UpdateQunTopicByIdSQL = "UPDATE `qun` SET `name` = ? WHERE `id` = ?"
	//根据群id和用户id删除群成员
	DelQunMemberByQunidAndUserid = "DELETE FROM `qun_user` WHERE `qun_id` =? AND `user_id` =?"
)

// 群结构.
type Qun struct {
	Id              string    `json:"id"`
	CreatorId       string    `json:"creatorId"`
	CreatorUserName string    `json:"creatorUserName"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	MaxMember       int       `json:"maxMember"`
	Avatar          string    `json:"avatar"`
	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
}

// 群-用户关联结构.
type QunUser struct {
	Id      string
	QunId   string
	UserId  string
	Sort    int
	Role    int
	Created time.Time
	Updated time.Time
}

// 创建群.
func (*device) CreateQun(w http.ResponseWriter, r *http.Request) {
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

	now := time.Now()

	creatorId := baseReq["uid"].(string)
	topic := args["topic"].(string)

	qid := uuid.New()
	qun := Qun{Id: qid, CreatorId: creatorId, Name: topic, Description: "", MaxMember: 100, Avatar: "", Created: now, Updated: now}

	memberList := args["memberList"].([]interface{})
	qunUsers := []QunUser{}
	for _, m := range memberList {
		member := m.(map[string]interface{})
		memberId := member["uid"].(string)

		if creatorId == memberId {
			// 创建者后面会单独处理
			continue
		}

		qunUser := QunUser{Id: uuid.New(), QunId: qid, UserId: memberId, Sort: 0, Role: 0, Created: now, Updated: now}

		qunUsers = append(qunUsers, qunUser)
	}

	creator := QunUser{Id: uuid.New(), QunId: qid, UserId: creatorId, Sort: 0, Role: 0, Created: now, Updated: now}
	qunUsers = append(qunUsers, creator)

	if createQun(&qun, qunUsers) {
		glog.Infof("Created Qun [id=%s]", qid)
	} else {
		glog.Error("Create Qun faild")
		baseRes.ErrMsg = "Create Qun faild"
		baseRes.Ret = InternalErr
	}

	res["ChatRoomName"] = qid + QUN_SUFFIX
	res["topic"] = topic
	res["memberCount"] = int(args["memberCount"].(float64))

	members, err := getUsersInQun(qid)
	if err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr
		return
	}

	res["memberList"] = members

	// 给创群人发送消息
	msg := map[string]interface{}{}
	msg["fromUserName"] = qid + QUN_SUFFIX
	msg["fromDisplayName"] = topic
	msg["msgType"] = 51
	msg["content"] = "你创建了群\"" + topic + "\""

	baseRes.Ret = pushSessions(msg, creatorId+USER_SUFFIX, []string{"all"}, 600)

	return
}

// 获取群成员.
func (*device) GetUsersInQun(w http.ResponseWriter, r *http.Request) {
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

	userName := args["userName"].(string)
	qid := userName[:strings.LastIndex(userName, QUN_SUFFIX)]

	members, err := getUsersInQun(qid)
	if err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr
		return
	}

	res["memberList"] = members
	res["memberCount"] = len(members)

	qun, err := getQunById(qid)
	if err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr
		return
	}
	qun.CreatorUserName = qun.CreatorId + USER_SUFFIX
	res["quninfo"] = qun

	return
}

//根据群id修改群topic
func (*device) UpdateQunTopicById(w http.ResponseWriter, r *http.Request) {
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
	chatRoomName := args["ChatRoomName"].(string)
	qunId := chatRoomName[:strings.LastIndex(chatRoomName, QUN_SUFFIX)]
	topic := args["topic"].(string)
	if updateQunTopicById(qunId, topic) {
		baseRes.ErrMsg = "update Qun Topic success"
		baseRes.Ret = OK

		msg := map[string]interface{}{}
		msg["fromUserName"] = qunId + QUN_SUFFIX
		msg["fromDisplayName"] = topic
		msg["toUserName"] = user.Uid + USER_SUFFIX
		msg["msgType"] = 51

		// 给修改者发送消息
		msg["content"] = "你修改了群名为\"" + topic + "\""
		pushSessions(msg, user.Uid+USER_SUFFIX, []string{"all"}, 600)

		//给其他群成员发送消息
		msg["content"] = user.NickName + "修改了群名为\"" + topic + "\""
		members, err := getUsersInQun(qunId)
		if err == nil {
			for _, mem := range members {
				if user.Uid == mem.Uid {
					continue
				}

				pushSessions(msg, mem.Uid+USER_SUFFIX, []string{"all"}, 600)
			}
		}
	} else {
		glog.Error("update Qun Topic  faild")
		baseRes.ErrMsg = "update Qun Topic  faild"
		baseRes.Ret = InternalErr
	}
}

//添加群成员
func (*device) AddQunMember(w http.ResponseWriter, r *http.Request) {
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
	chatRoomName := args["ChatRoomName"].(string)
	qunId := chatRoomName[:strings.LastIndex(chatRoomName, QUN_SUFFIX)]
	qun, err := getQunById(qunId)
	if err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr
		return
	}
	qun.CreatorUserName = qun.CreatorId + USER_SUFFIX
	res["quninfo"] = qun

	memberList := args["memberList"].([]interface{})
	qunUsers := []QunUser{}
	now := time.Now()
	for _, m := range memberList {
		member := m.(map[string]interface{})
		memberId := member["uid"].(string)
		qunUser := QunUser{Id: uuid.New(), QunId: qunId, UserId: memberId, Sort: 0, Role: 0, Created: now, Updated: now}
		qunUsers = append(qunUsers, qunUser)
	}
	if addQunmember(qunUsers) {
		members, err := getUsersInQun(qunId)
		if err != nil {
			baseRes.ErrMsg = err.Error()
			baseRes.Ret = InternalErr
			glog.Error("get Qun Member faild")
			return
		}
		res["memberList"] = members
		res["memberCount"] = len(members)
	} else {
		glog.Error("add Qun Member   faild")
		baseRes.ErrMsg = "add Qun Member faild"
		baseRes.Ret = InternalErr
	}
}

//删除群成员
func (*device) DelQunMember(w http.ResponseWriter, r *http.Request) {
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
	chatRoomName := args["ChatRoomName"].(string)
	qunId := chatRoomName[:strings.LastIndex(chatRoomName, QUN_SUFFIX)]
	qun, err := getQunById(qunId)
	if err != nil {
		baseRes.ErrMsg = err.Error()
		baseRes.Ret = InternalErr
		return
	}
	qun.CreatorUserName = qun.CreatorId + USER_SUFFIX
	res["quninfo"] = qun
	creatorId := qun.CreatorId
	if creatorId != user.Uid {
		glog.Error("delete Qun Member   faild,only qun creater can delete")
		baseRes.ErrMsg = "delete Qun Member faild,only qun creater can delete"
		baseRes.Ret = InternalErr
		return
	}
	memberList := args["memberList"].([]interface{})
	qunUsers := []QunUser{}
	for _, m := range memberList {
		member := m.(map[string]interface{})
		memberId := member["uid"].(string)
		qunUser := QunUser{QunId: qunId, UserId: memberId}
		qunUsers = append(qunUsers, qunUser)
	}
	if DelQunMember(qunUsers) {
		members, err := getUsersInQun(qunId)
		if err != nil {
			baseRes.ErrMsg = err.Error()
			baseRes.Ret = InternalErr
			glog.Error("get Qun Member faild")
			return
		}
		res["memberList"] = members
		res["memberCount"] = len(members)
	} else {
		glog.Error("delete Qun Member   faild")
		baseRes.ErrMsg = "delete Qun Member faild"
		baseRes.Ret = InternalErr
	}
}

//根据群id更新群topic
func updateQunTopicById(qunId string, topic string) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)

		return false
	}
	_, err = tx.Exec(UpdateQunTopicByIdSQL, topic, qunId)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}

		return false
	}
	if err = tx.Commit(); err != nil {
		glog.Error(err)

		return false
	}
	return true
}

// 数据库中插入群记录、群-用户关联记录.
func createQun(qun *Qun, qunUsers []QunUser) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)

		return false
	}

	// 创建群记录
	_, err = tx.Exec(InsertQunSQL, qun.Id, qun.CreatorId, qun.Name, qun.Description, qun.MaxMember, qun.Avatar, qun.Created, qun.Updated)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}

		return false
	}

	// 创建群成员关联
	for _, qunUser := range qunUsers {
		_, err = tx.Exec(InsertQunUserSQL, qunUser.Id, qunUser.QunId, qunUser.UserId, qunUser.Sort, qunUser.Role, qunUser.Created, qunUser.Updated, qunUser.QunId, qunUser.UserId)

		if err != nil {
			glog.Error(err)

			if err := tx.Rollback(); err != nil {
				glog.Error(err)
			}

			return false
		}
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)

		return false
	}

	return true
}

//添加群成员
func addQunmember(qunUsers []QunUser) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)

		return false
	}

	for _, qunUser := range qunUsers {
		_, err = tx.Exec(InsertQunUserSQL, qunUser.Id, qunUser.QunId, qunUser.UserId, qunUser.Sort, qunUser.Role, qunUser.Created, qunUser.Updated, qunUser.QunId, qunUser.UserId)
		if err != nil {
			glog.Error(err)

			if err := tx.Rollback(); err != nil {
				glog.Error(err)
			}

			return false
		}
	}
	if err := tx.Commit(); err != nil {
		glog.Error(err)

		return false
	}

	return true

}

//删除群成员
func DelQunMember(qunUsers []QunUser) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)

		return false
	}

	for _, qunUser := range qunUsers {
		_, err = tx.Exec(DelQunMemberByQunidAndUserid, qunUser.QunId, qunUser.UserId)
		if err != nil {
			glog.Error(err)

			if err := tx.Rollback(); err != nil {
				glog.Error(err)
			}

			return false
		}
	}
	if err := tx.Commit(); err != nil {
		glog.Error(err)

		return false
	}

	return true
}

// 在数据库中查询群内用户.
func getUsersInQun(qunId string) ([]member, error) {
	ret := []member{}
	glog.Infoln("qunId", qunId)
	rows, err := db.MySQL.Query(SelectQunUserSQL, qunId)
	if err != nil {
		glog.Error(err)

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		rec := member{}

		if err := rows.Scan(&rec.Uid, &rec.Name, &rec.NickName, &rec.Status, &rec.Avatar, &rec.TenantId, &rec.PYInitial, &rec.PYQuanPin, &rec.Mobile, &rec.Area); err != nil {
			glog.Error(err)

			return nil, err
		}

		rec.UserName = rec.Uid + USER_SUFFIX

		ret = append(ret, rec)
	}

	if err := rows.Err(); err != nil {
		glog.Error(err)

		return nil, err
	}

	return ret, nil
}

// 在数据库中查询群内用户 id.
func getUserIdsInQun(qunId string) ([]string, error) {
	ret := []string{}

	rows, err := db.MySQL.Query(SelectQunUserIdSQL, qunId)
	if err != nil {
		glog.Error(err)

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid string

		if err := rows.Scan(&uid); err != nil {
			glog.Error(err)

			return nil, err
		}

		ret = append(ret, uid)
	}

	if err := rows.Err(); err != nil {
		glog.Error(err)

		return nil, err
	}

	return ret, nil
}

// 在数据库中查询群.
func getQunById(qunId string) (*Qun, error) {
	row := db.MySQL.QueryRow(SelectQunById, qunId)

	qun := Qun{}
	if err := row.Scan(&qun.Id, &qun.CreatorId, &qun.Name, &qun.Description, &qun.MaxMember, &qun.Avatar, &qun.Created, &qun.Updated); err != nil {
		glog.Error(err)

		return nil, err
	}

	return &qun, nil
}
