package session

import (
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/golang/glog"
	"time"
)

const (
	SESSION_STATE_ACTIVE    = "active"
	SESSION_STATE_INACTIVE  = "inactive"
	INSERT_SESSION          = "INSERT INTO `session`(`id`,`type`,`user_id`,`state`,`created`,`updated`)  VALUES (?,?,?,?,?,?) "
	DELETE_SESSION_BYID     = "DELETE FROM `session` WHERE `id`=?"
	DELETE_SESSION_BYUSERID = "DELETE FROM `session` WHERE `user_id`=?"
	UPDATE_SESSION_UPDATED  = "UPDATE `session` SET `updated` = ?  WHERE `id`=? "
	SELECT_SESSION_ALL      = "SELECT  `id`,`type`,`user_id`,`created`,`updated` FROM SESSION"
	DELETE_SESSION_PAST     = "DELETE  FROM `session` WHERE  `updated` < ?"
	SET_USERID              = "UPDATE `session` SET  `user_id` = ? WHERE  `id` = ?"
)

type Session struct {
	Id      string    `json:"id"`
	Type    string    `json:"type"`
	UserId  string    `json:"userId"`
	Sate    string    `json:"sate"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

//一个星期扫描一次
var ScanSessionTime = time.NewTicker(168 * time.Hour)

//创建会话session记录
func CreatSession(session *Session) bool {

	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	//var e error
	_, err = tx.Exec(INSERT_SESSION, session.Id, session.Type, session.UserId, session.Sate, session.Created, session.Updated)
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

//修改会话session记录
func UpdateSessionUserID(sessionId, userId string) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	_, err = tx.Exec(SET_USERID, sessionId, userId)
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

///除过时的会话,当更新时间距当前时间超过24 小时，就将该会话移除
func ScanSession() {

	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
	}
	for t := range ScanSessionTime.C {

		glog.Info(t.Format("2006/01/02 15:04:05"), "执行了定时扫描session表操作-----------------------------")
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
