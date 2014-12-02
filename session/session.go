package session

import (
	"database/sql"
	"strings"
	"time"

	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/golang/glog"
)

const (
	SESSION_STATE_INIT     = "init"
	SESSION_STATE_ACTIVE   = "active"
	SESSION_STATE_INACTIVE = "inactive"

	INSERT_SESSION          = "INSERT INTO `session`(`id`,`type`,`user_id`,`state`,`created`,`updated`)  VALUES (?,?,?,?,?,?) "
	DELETE_SESSION_BYID     = "DELETE FROM `session` WHERE `id`=?"
	DELETE_SESSION_BYUSERID = "DELETE FROM `session` WHERE `user_id`=?"
	UPDATE_SESSION_UPDATED  = "UPDATE `session` SET `updated` = ?  WHERE `id`=? "
	UPDATE_SESSION_STATE    = "UPDATE `session` SET `state` = ?  WHERE `id`=? "
	DELETE_SESSION_PAST     = "DELETE  FROM `session` WHERE  `updated` < ?"
	SET_USERID              = "UPDATE `session` SET  `user_id` = ? WHERE  `id` = ?"
	SELECT_SESSION_BYSTATE  = "SELECT  `id`,`type`,`user_id`,`state`,`created`,`updated` FROM SESSION WHERE `user_id`=?  AND `state`=?"
	SELECT_SESSION_BYUSERID = "SELECT  `id`,`type`,`user_id`,`state`,`created`,`updated` FROM SESSION WHERE `user_id` = ?"
	SELECT_SESSION_BYID     = "SELECT  `id`,`type`,`user_id`,`state`,`created`,`updated` FROM SESSION WHERE `id`=?"
	SELECT_SESSION_BYIDS    = "SELECT  `id`,`type`,`user_id`,`state`,`created`,`updated` FROM SESSION WHERE `id`IN (？)"
	EXIST_SESSION           = "SELECT id FROM session WHERE id = ?"
)

type Session struct {
	Id      string    `json:"id"`
	Type    string    `json:"type"`
	UserId  string    `json:"userId"`
	State   string    `json:"sate"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

//一个星期扫描一次
var ScanSessionTime = time.NewTicker(168 * time.Hour)

// 根据 args 参数获取用户 uid 的会话集.
//
// args 参数：
//  1. ["all"] 表示获取该用户所有的会话
//  2. ["xxx1", "xxx2"] 表示获取该用户 xxx1、xxx2 会话
//  3. ["active"] 表示获取该用户的所有激活的会话
//  4. [inactive"] 表示获取该用户的所有非激活的会话
func GetSessions(uid string, args []string) []*Session {
	ret := []*Session{}

	length := len(args)
	if 0 == length {
		return ret
	}
	var rows *sql.Rows
	var err error
	if 1 == length {

		first := args[0]

		switch first {
		case "all":
			rows, err = db.MySQL.Query(SELECT_SESSION_BYUSERID, uid)
		case "active":
			rows, err = db.MySQL.Query(SELECT_SESSION_BYSTATE, uid, "active")
		case "inactive":
			rows, err = db.MySQL.Query(SELECT_SESSION_BYSTATE, uid, "inactive")
		default: // 只有 1 个会话的情况
			rows, err = db.MySQL.Query(SELECT_SESSION_BYID, first)
		}
	} else {
		// 大于 1 一个会话的情况都是指定到了具体的会话 id
		rows, err = db.MySQL.Query(SELECT_SESSION_BYIDS, strings.Join(args, ","))
	}
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		glog.Error(err)
		return ret
	}
	if err := rows.Err(); err != nil {
		glog.Error(err)
		return ret
	}
	for rows.Next() {
		session := &Session{}
		if err := rows.Scan(&session.Id, &session.Type, &session.UserId, &session.State, &session.Created, &session.Updated); err != nil {
			glog.Error(err)
			return ret
		}
		ret = append(ret, session)
	}

	return ret
}

//创建会话session记录
func CreatSession(session *Session) bool {
	//存在不做任何处理
	if IsExistSession(session.Id) {
		return true
	}
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}

	_, err = tx.Exec(INSERT_SESSION, session.Id, session.Type, session.UserId, session.State, session.Created, session.Updated)
	if err != nil {
		if !strings.Contains(err.Error(), "Duplicate entry") {
			// 非主键重复的异常才打日志
			glog.Error(err)
		}

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}

		return false
	}

	//提交操作
	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}
	return true
}

func IsExistSession(sessionId string) bool {
	rows, err := db.MySQL.Query(EXIST_SESSION, sessionId)
	if err != nil {
		glog.Error(err)
		return false
	}

	defer rows.Close()

	if err = rows.Err(); err != nil {
		glog.Error(err)
		return false
	}
	return rows.Next()
}

//修改会话session记录
func UpdateSessionUserID(sessionId, userId string) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(SET_USERID, userId, sessionId)
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}
	//提交操作
	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}
	return true
}

//删除会话session记录
func RemoveSessionById(id string) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}
	_, err = tx.Exec(DELETE_SESSION_BYID, id)
	if err != nil {
		glog.Error(err)
		return false
	}
	//提交操作
	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}
	return true
}

//删除会话session记录
func RemoveSessionByUserId(usreId string) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(DELETE_SESSION_BYUSERID, usreId)
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}
	//提交操作
	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}
	return true
}

//更新会话session的最新时间
func UpdateSessionUpdated(sessionId string, updated time.Time) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(UPDATE_SESSION_UPDATED, updated, sessionId)
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}
	//提交操作
	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}
	return true
}

//定时更新会话session时间
func TickerTaskUpdateSession(sessionId string, tickerFlagStop chan bool) {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			UpdateSessionUpdated(sessionId, time.Now().Local())
		case <-tickerFlagStop:
			return
		}
	}
}

//更新会话状态
func SetSessionStat(sessionId, state string) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(UPDATE_SESSION_STATE, state, sessionId)
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}
	//提交操作
	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return false
	}
	return true
}

//通过userd查询会话session
func GetSessionsByUserId(userId string) (*[]Session, error) {

	rows, err := db.MySQL.Query(SELECT_SESSION_BYUSERID, userId)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	sessions := []Session{}
	for rows.Next() {
		session := Session{}
		if err := rows.Scan(&session.Id, &session.Type, &session.UserId, &session.State, &session.Created, &session.Updated); err != nil {
			glog.Error(err)
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		glog.Error(err)
		return nil, err
	}
	return &sessions, nil
}

///除过时的会话,当更新时间距当前时间超过24 小时，就将该会话移除
func ScanSession() {

	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
	}
	for t := range ScanSessionTime.C {

		glog.V(3).Info(t.Format("2006/01/02 15:04:05"), "执行了定时扫描session表操作-----------------------------")
		//创建过时时间
		hours, _ := time.ParseDuration("-24h")
		pastTime := time.Now().Local()
		pastTime = pastTime.Add(hours)

		_, err = tx.Exec(DELETE_SESSION_PAST, pastTime)
		if err != nil {
			glog.Error(err)
			if err := tx.Rollback(); err != nil {
				glog.Error(err)
			}
		}
		//提交操作
		if err := tx.Commit(); err != nil {
			glog.Error(err)
		}
	}
}
