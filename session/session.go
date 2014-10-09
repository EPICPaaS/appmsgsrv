package session

import (
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/golang/glog"
	"time"
)

const (
	SESSION_STATE_ACTIVE   = "active"
	SESSION_STATE_INACTIVE = "inactive"
	INSERTSESSION          = "INSERT INTO 'session'('id','type','user_id','state','created','updated')  VALUES (?,?,?,?,?,?) "
)

type Session struct {
	Id      string    `json:"id"`
	Type    string    `json:"type"`
	UserId  string    `json:"userId"`
	Sate    string    `json:"sate"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func CreatSession(session *Session) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	//""INSERT INTO 'session'('id','type','user_id','state','created','updated')  VALUES (?,?,?,?,?,?) ""
	_, err = tx.Exec(INSERTSESSION, session.Id, session.Type, session.UserId, session.Sate, session.Created, session.Updated)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}

	return true
}
