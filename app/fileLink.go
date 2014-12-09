package app

import (
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
	"time"
)

const (
	INSERT_FILELINK      = "insert into file_link (id , sender_id,file_id ,file_name,file_url,size,created,updated)  values(?,?,?,?,?,?,?,?)"
	UPDATE_FILELINK_TIME = "update file_link set updated =? where sender_id =? and file_id =?"
	EXIST_FILELINK       = "select id from file_link where sender_id =? and file_id =?"
)

type FileLink struct {
	Id       string
	SenderId string
	FileId   string
	FileName string
	FileUrl  string
	Size     int
	Created  time.Time
	Updated  time.Time
}

func SaveFileLinK(fileLink *FileLink) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		glog.Error(err)
		return false
	}
	//更新
	if ExistFileLink(fileLink) {
		_, err = tx.Exec(UPDATE_FILELINK_TIME, time.Now().Local(), fileLink.SenderId, fileLink.FileId)
	} else { //新增
		_, err = tx.Exec(INSERT_FILELINK, uuid.New(), fileLink.SenderId, fileLink.FileId, fileLink.FileName, fileLink.FileUrl, fileLink.Size, time.Now().Local(), time.Now().Local())
	}
	if err != nil {
		glog.Error(err)
		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return false
	}

	if err := tx.Commit(); err != nil {
		return false
	}

	return true
}

func ExistFileLink(fileLink *FileLink) bool {
	rows, err := db.MySQL.Query(EXIST_FILELINK, fileLink.SenderId, fileLink.FileId)
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
