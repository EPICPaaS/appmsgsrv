package app

import (
	"encoding/json"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/go-uuid/uuid"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	INSERT_FILELINK        = "insert into file_link (id , sender_id,file_id ,file_name,file_url,size,created,updated)  values(?,?,?,?,?,?,?,?)"
	UPDATE_FILELINK_TIME   = "update file_link set updated =? where sender_id =? and file_id =?"
	EXIST_FILELINK         = "select id from file_link where sender_id =? and file_id =?"
	SELECT_EXPIRE_FILELINK = "select  id, file_url from file_link where  updated  < ?"
)

var ScanFileTime = time.NewTicker(5 * time.Minute)

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

/*保存文件链接信息*/
func SaveFileLinK(fileLink *FileLink) bool {
	tx, err := db.MySQL.Begin()
	if err != nil {
		logger.Error(err)
		return false
	}
	//更新
	if ExistFileLink(fileLink) {
		_, err = tx.Exec(UPDATE_FILELINK_TIME, time.Now().Local(), fileLink.SenderId, fileLink.FileId)
	} else { //新增
		_, err = tx.Exec(INSERT_FILELINK, uuid.New(), fileLink.SenderId, fileLink.FileId, fileLink.FileName, fileLink.FileUrl, fileLink.Size, time.Now().Local(), time.Now().Local())
	}
	if err != nil {
		logger.Error(err)
		if err := tx.Rollback(); err != nil {
			logger.Error(err)
		}
		return false
	}

	if err := tx.Commit(); err != nil {
		return false
	}

	return true
}

/*判断是否存在文件链接记录*/
func ExistFileLink(fileLink *FileLink) bool {
	rows, err := db.MySQL.Query(EXIST_FILELINK, fileLink.SenderId, fileLink.FileId)
	if err != nil {
		logger.Error(err)
		return false
	}

	defer rows.Close()

	if err = rows.Err(); err != nil {
		logger.Error(err)
		return false
	}
	return rows.Next()
}

/*删除weedfs服务器文件*/
func DeleteFile(fileUrl string) bool {

	var url = "http://" + fileUrl
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		logger.Errorf("delete file fail  [ERROR]-%s", err.Error())
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("delete file fail  [ERROR]-%s", err.Error())
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return false
	}
	var respBody map[string]interface{}
	if err := json.Unmarshal(body, &respBody); err != nil {
		logger.Errorf("convert to json failed (%s)", err.Error())
		return false
	}
	e, ok := respBody["error"].(string)
	if ok {
		logger.Errorf("delete file fail [ERROR]- %s", e)
		return false
	}
	return true
}

/*定时扫描过期的文件链接，如果过期侧删除该文件记录和文件服务器中的文件*/
func ScanExpireFileLink() {

	//构造时间差
	subTimeStr := strconv.Itoa(Conf.MsgExpire)
	subTime, _ := time.ParseDuration("-" + subTimeStr + "s")
	/*定时任务删除，过期聊天文件*/
	for t := range ScanFileTime.C {

		expire := time.Now().Local().Add(subTime)

		rows, err := db.MySQL.Query(SELECT_EXPIRE_FILELINK, expire)
		if err != nil {
			logger.Error(err)
			return
		}

		defer rows.Close()
		if err := rows.Err(); err != nil {
			logger.Error(err)
			return
		}

		var delIds []string
		for rows.Next() {
			var id, fileUrl string
			if err := rows.Scan(&id, &fileUrl); err != nil {
				logger.Error(err)
				continue
			}

			//删除文件
			if DeleteFile(fileUrl) {
				delIds = append(delIds, id)
			}
		}

		/*删除文件记录*/
		if len(delIds) > 0 {
			tx, err := db.MySQL.Begin()
			if err != nil {
				logger.Error(err)
				return
			}
			delSql := "delete  from file_link where  id   in ('" + strings.Join(delIds, "','") + "')"
			_, err = tx.Exec(delSql)
			if err != nil {
				logger.Error(err)
				if err := tx.Rollback(); err != nil {
					logger.Error(err)
				}
				return
			}
			//提交操作
			if err := tx.Commit(); err != nil {
				logger.Error(err)
				return
			}
		}
		logger.Infof("%v scan file succeed", t)
	}
	logger.Error("scan expire  file logout ")
}
