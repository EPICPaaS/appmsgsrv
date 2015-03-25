package app

import (
	"database/sql"
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"io/ioutil"
	"net/http"
	"time"
)

type ShieldMsg struct {
	Id       string    `json:"id"`
	SourceId string    `json:"sourceId"`
	TargetId string    `json:"targetId"`
	Memo     string    `json:"memo"`
	Type     int       `json:"type"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

const (
	SELECT_SHIELDMSG_BY_SOURCE_TYPE  = "select id , source_id,target_id,memo,type,created,updated from shield_msg where source_id = ? and type =?"
	SELECT_SHIELDMSG_BY_SOURCE       = "select id , source_id,target_id,memo,type,created,updated from shield_msg where source_id = ? "
	SELECT_SHIELDMSG_BY_UTARGET_TYPE = "select id , source_id,target_id,memo,type,created,updated from shield_msg where target_id = ? and type =?"
	SELECT_SHIELDMSG_BY_UTARGET      = "select id , source_id,target_id,memo,type,created,updated from shield_msg where target_id = ? "
	INSERT_SHIELDMSG                 = "insert into shield_msg(id , source_id,target_id,memo,type,created,updated) values(?,?,?,?,?,?,?)"
	DELETE_SHIELDMSG                 = "delete from shield_msg where source_id = ? and target_id = ?  and type =? "
)

/*获取我屏蔽里谁
@uid为用户id
@shieldType 为屏蔽类型（用户消息/应用消息）shieldType 为空(零值)时为回去有所类型屏蔽记录
*/
func GetShieldSourceMsg(sourceId string, shieldType int) (shieldMsgs []ShieldMsg) {

	var rows *sql.Rows
	var err error

	if shieldType == 0 {
		rows, err = db.MySQL.Query(SELECT_SHIELDMSG_BY_SOURCE_TYPE, sourceId, shieldType)
	} else { //零值查询所有屏蔽类型
		rows, err = db.MySQL.Query(SELECT_SHIELDMSG_BY_SOURCE, sourceId)
	}

	if err == nil {
		defer rows.Close()
	} else {
		logger.Error(err)
		return nil
	}
	if err := rows.Err(); err != nil {
		logger.Error(err)
		return nil
	}

	for rows.Next() {
		shieldMsg := ShieldMsg{}
		if err := rows.Scan(&shieldMsg.Id, &shieldMsg.SourceId, &shieldMsg.TargetId, &shieldMsg.Memo, &shieldMsg.Type, &shieldMsg.Created, &shieldMsg.Updated); err != nil {
			logger.Error(err)
			continue
		}
		shieldMsgs = append(shieldMsgs, shieldMsg)
	}

	return
}

/*获取谁屏蔽里我
@uid为用户id
@shieldType 为屏蔽类型（用户消息/应用消息）shieldType 为空(零值)时为回去有所类型屏蔽记录
*/
func GetShieldTargetMsg(targetId string, shieldType int) (shieldMsgs []ShieldMsg) {

	var rows *sql.Rows
	var err error

	if shieldType == 0 {
		rows, err = db.MySQL.Query(SELECT_SHIELDMSG_BY_UTARGET_TYPE, targetId, shieldType)
	} else { //零值查询所有屏蔽类型
		rows, err = db.MySQL.Query(SELECT_SHIELDMSG_BY_UTARGET, targetId)
	}

	if err == nil {
		defer rows.Close()
	} else {
		logger.Error(err)
		return nil
	}
	if err := rows.Err(); err != nil {
		logger.Error(err)
		return nil
	}

	for rows.Next() {
		shieldMsg := ShieldMsg{}
		if err := rows.Scan(&shieldMsg.Id, &shieldMsg.SourceId, &shieldMsg.TargetId, &shieldMsg.Memo, &shieldMsg.Type, &shieldMsg.Created, &shieldMsg.Updated); err != nil {
			logger.Error(err)
			continue
		}
		shieldMsgs = append(shieldMsgs, shieldMsg)
	}

	return
}

/*判断限制是否存在*/
func isExistShieldMsg(sourceId, targetId string, shieldtype int) bool {

	rows, err := db.MySQL.Query("select id from shield_msg where source_id =? and target_id =? and type = ?", sourceId, targetId, shieldtype)

	if err != nil {
		logger.Error(err)
		return false
	}

	defer rows.Close()
	if err := rows.Err(); err != nil {
		logger.Error(err)
		return false
	}

	return rows.Next()
}

/*设置屏蔽消息*/
func (*device) SetShieldMsg(w http.ResponseWriter, r *http.Request) {

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

	sourceId := args["sourceId"].(string)
	targetId := args["targetId"].(string)
	shieldtype := int(args["type"].(float64))
	memo, _ := args["memo"].(string)

	shieldMsg := &ShieldMsg{
		SourceId: sourceId,
		TargetId: targetId,
		Type:     shieldtype,
		Memo:     memo,
	}

	if !SaveShieldMsg(shieldMsg) {
		baseRes.Ret = InternalErr
		return
	}
	return
}

/*设置屏蔽消息*/
func (*device) IsUserShieldApp(w http.ResponseWriter, r *http.Request) {

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

	sourceId := args["sourceId"].(string)
	targetId := args["targetId"].(string)
	res["isShield"] = isExistShieldMsg(sourceId, targetId, 1)

	return
}

/*取消屏蔽消息*/
func (*device) CancelShieldMsg(w http.ResponseWriter, r *http.Request) {

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

	sourceId := args["sourceId"].(string)
	targetId := args["targetId"].(string)
	shieldtype := int(args["type"].(float64))

	shieldMsg := &ShieldMsg{
		SourceId: sourceId,
		TargetId: targetId,
		Type:     shieldtype,
	}

	if !DelShieldMsg(shieldMsg) {
		baseRes.Ret = InternalErr
		return
	}
	return
}

/*保存屏蔽记录*/
func SaveShieldMsg(shieldMsg *ShieldMsg) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}
	_, err = tx.Exec(INSERT_SHIELDMSG, uuid.New(), shieldMsg.SourceId, shieldMsg.TargetId, shieldMsg.Memo, shieldMsg.Type, time.Now().Local(), time.Now().Local())
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

/*删除屏蔽消息记录*/
func DelShieldMsg(shieldMsg *ShieldMsg) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}
	_, err = tx.Exec(DELETE_SHIELDMSG, shieldMsg.SourceId, shieldMsg.TargetId, shieldMsg.Type)
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
