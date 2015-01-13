package app

import (
	"database/sql"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/EPICPaaS/yixinappsrv/db"
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
