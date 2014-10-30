package app

import (
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// 添加联系人.
	InsertUserUserSQL = "INSERT INTO `user_user` (`id`, `from_user_id`, `to_user_id`, `remark_name`, `sort`,`created`, `updated`) VALUES " +
		"(?, ?, ?, ?, ?, ?, ?)"
	// 删除联系人.
	DeleteUserUserSQL = "DELETE FROM `user_user` WHERE `from_user_id` = ? AND `to_user_id` = ?"
)

// 联系人结构.
type UserUser struct {
	Id         string    `json:"id"`
	FromUserId string    `json:"fromUserId"`
	ToUserId   string    `json:"toUserId"`
	RemarkName string    `json:"remarkName"`
	Sort       int       `json:"sort"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
}

/*添加或删除联系人 ， 当请求参数starFriend为1时添加联系人，否则删除*/
func (*device) AddOrRemoveContact(w http.ResponseWriter, r *http.Request) {
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

	fromUserId := baseReq["uid"].(string)
	toUserId := args["uid"].(string)
	starFriend := int(args["starFriend"].(float64))

	now := time.Now()

	if 1 == starFriend { // 添加联系人
		if isStar(fromUserId, toUserId) {
			return
		}

		userUser := UserUser{Id: uuid.New(), FromUserId: fromUserId, ToUserId: toUserId, RemarkName: "", Sort: 0,
			Created: now, Updated: now}

		if !createContact(&userUser) {
			baseRes.Ret = InternalErr

			return
		}

		glog.V(3).Infof("Created a contact [from=%s, to=%s]", fromUserId, toUserId)
	} else { // 删除联系人
		if !deleteContact(fromUserId, toUserId) {
			baseRes.Ret = InternalErr

			return
		}

		glog.V(3).Infof("Deleted a contact [from=%s, to=%s]", fromUserId, toUserId)
	}
}

// 数据库中插入联系人记录.
func createContact(userUser *UserUser) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)

		return false
	}

	_, err = tx.Exec(InsertUserUserSQL, userUser.Id, userUser.FromUserId, userUser.ToUserId, userUser.RemarkName,
		userUser.Sort, userUser.Created, userUser.Updated)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}

		return false
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)

		return false
	}

	return true
}

// 在数据库中删除联系人记录.
func deleteContact(fromUserId, toUserId string) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)

		return false
	}

	_, err = tx.Exec(DeleteUserUserSQL, fromUserId, toUserId)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}

		return false
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)

		return false
	}

	return true
}
