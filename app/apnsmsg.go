package app

import (
	"time"

	"github.com/EPICPaaS/appmsgsrv/db"
	//"github.com/EPICPaaS/go-uuid/uuid"
)

const (

	// 租户资源插入 SQL.
	InsertApnsMsgSQL = "INSERT INTO `apns_msg` (`id`, `type`, `apns_token`, `cnt`, `tenant_id`, `pushed`, `created`, `updated`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	// 根据资源id 获取资源
	SelectApnsMsgByIdSQL = "SELECT * FROM `apns_msg` where `id` = ?"
	// 根据租户 id 查询租户的资源.
	SelectApnsMsgByTenantIdSQL = "SELECT * FROM `apns_msg` where `tenant_id`  = ?"
	// 根据apns_token 查询租户的资源.
	SelectApnsMsgByApnsTokenSQL = "SELECT * FROM `apns_msg` where `apns_token`  = ?"
	//根据id修改资源
	UpdateApnsMsgByIdSQL = "UPDATE `apns_msg` SET `type` = ? ,`apns_token` = ? , `cnt` = ? , `tenant_id` = ? ,`pushed` = ? ,  `created` = ?  , `updated` = ?   WHERE `id` = ?"
	//根据id删除资源
	DelApnsMsgByIdSQL = "DELETE FROM `apns_msg` WHERE `id` =? "
	//根据tenant_id删除资源
	DelApnsMsgByTenantIdSQL = "DELETE FROM `apns_msg` where `tenant_id`  = ?"
)

/**
APNS消息
**/
type ApnsMsg struct {
	Id        string    `json:"id"`
	Type      string    `json:"type"`
	ApnsToken string    `json:"apns_token"`
	Cnt       int       `json:"cnt"`
	TenantId  string    `json:"tenant_id"`
	Pushed    string    `json:"pushed"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

// 在数据库中查询资源.
func GetApnsMsgById(id string) (*ApnsMsg, error) {
	row := db.MySQL.QueryRow(SelectApnsMsgByIdSQL, id)

	apnsmsg := ApnsMsg{}
	if err := row.Scan(&apnsmsg.Id, &apnsmsg.Type, &apnsmsg.ApnsToken, &apnsmsg.Cnt, &apnsmsg.TenantId, &apnsmsg.Pushed, &apnsmsg.Created, &apnsmsg.Updated); err != nil {
		logger.Error(err)

		return nil, err
	}

	return &apnsmsg, nil
}

// 在数据库中查询资源.
func GetApnsMsgByTenantId(tenantId string) ([]*ApnsMsg, error) {

	rows, _ := db.MySQL.Query(SelectApnsMsgByTenantIdSQL, tenantId)

	ret := []*ApnsMsg{}
	for rows.Next() {
		apnsmsg := ApnsMsg{}
		if err := rows.Scan(&apnsmsg.Id, &apnsmsg.Type, &apnsmsg.ApnsToken, &apnsmsg.Cnt, &apnsmsg.TenantId, &apnsmsg.Pushed, &apnsmsg.Created, &apnsmsg.Updated); err != nil {
			logger.Error(err)

			return nil, err
		}
		ret = append(ret, &apnsmsg)
	}

	return ret, nil
}

// 在数据库中查询资源.
func GetApnsMsgByApnsToken(apnsToken string) ([]*ApnsMsg, error) {

	rows, _ := db.MySQL.Query(SelectApnsMsgByApnsTokenSQL, apnsToken)

	ret := []*ApnsMsg{}
	for rows.Next() {
		apnsmsg := ApnsMsg{}
		if err := rows.Scan(&apnsmsg.Id, &apnsmsg.Type, &apnsmsg.ApnsToken, &apnsmsg.Cnt, &apnsmsg.TenantId, &apnsmsg.Pushed, &apnsmsg.Created, &apnsmsg.Updated); err != nil {
			logger.Error(err)

			return nil, err
		}
		ret = append(ret, &apnsmsg)
	}

	return ret, nil
}

// 数据库中插入资源
func AddApnsMsg(apnsMsg *ApnsMsg) (*ApnsMsg, bool) {
	tx, err := db.MySQL.Begin()

	if err != nil {
		logger.Error(err)
		return nil, false
	}

	// 创建资源记录
	_, err = tx.Exec(InsertResourceSQL, apnsMsg.Id, apnsMsg.Type, apnsMsg.ApnsToken, apnsMsg.Cnt, apnsMsg.TenantId, apnsMsg.Pushed, apnsMsg.Created, apnsMsg.Updated)
	if err != nil {
		logger.Error(err)

		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return nil, false
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return nil, false
	}
	return apnsMsg, true
}

// 数据库中插入资源
func UpdateApnsMsgById(apnsMsg *ApnsMsg) (*ApnsMsg, bool) {
	tx, err := db.MySQL.Begin()

	if err != nil {
		logger.Error(err)
		return nil, false
	}

	// 创建资源记录
	_, err = tx.Exec(UpdateApnsMsgByIdSQL, apnsMsg.Type, apnsMsg.ApnsToken, apnsMsg.Cnt, apnsMsg.TenantId, apnsMsg.Pushed, apnsMsg.Created, apnsMsg.Updated, apnsMsg.Id)
	if err != nil {
		logger.Error(err)

		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return nil, false
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return nil, false
	}
	return apnsMsg, true
}

func DeleteApnsMsgById(id string) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		logger.Error(err)

		return false
	}

	_, err = tx.Exec(DelApnsMsgByIdSQL, id)
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

func DeleteApnsMsgByTenantId(tenantId string) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		logger.Error(err)

		return false
	}

	_, err = tx.Exec(DelApnsMsgByTenantIdSQL, tenantId)
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
