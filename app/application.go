package app

import (
	//"database/sql"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/golang/glog"
	"time"
)

const (
	// 根据 id 查询应用记录.
	SelectApplicationById = "SELECT * FROM `application` WHERE `id` = ?"
	// 查询应用记录.
	SelectAllApplication = "SELECT * FROM `application`"
	// 根据 token 获取应用记录.
	SelectApplicationByToken = "SELECT * FROM `application` WHERE `token` = ?"
)

// 应用结构.
type application struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Token    string    `json:"token"`
	Type     string    `json:"type"`
	Status   int       `json:"status"`
	Sort     int       `json:"sort"`
	Level    int       `json:"level"`
	Icon     string    `json:"icon"`
	TenantId string    `json:tenantId`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// 根据 id 查询应用记录.
func getApplication(appId string) (*application, error) {
	row := db.MySQL.QueryRow(SelectApplicationById, appId)

	application := application{}

	if err := row.Scan(&application.Id, &application.Name, &application.Token, &application.Type, &application.Status,
		&application.Sort, &application.Level, &application.Icon, &application.Created, &application.Updated); err != nil {
		glog.Error(err)

		return nil, err
	}

	return &application, nil
}

func getAllApplication() ([]*application, error) {
	rows, _ := db.MySQL.Query(SelectAllApplication)
	if rows != nil {
		defer rows.Close()
	}
	ret := []*application{}
	for rows.Next() {
		application := &application{}
		if err := rows.Scan(&application.Id, &application.Name, &application.Token, &application.Type, &application.Status,
			&application.Sort, &application.Level, &application.Icon, &application.Created, &application.Updated); err != nil {
			glog.Error(err)

			return nil, err
		}
		//去处Token值，防止暴露
		application.Token = ""
		ret = append(ret, application)
	}

	return ret, nil
}

// 根据 token 查询应用记录.
func getApplicationByToken(token string) (*application, error) {
	row := db.MySQL.QueryRow(SelectApplicationByToken, token)

	application := application{}

	if err := row.Scan(&application.Id, &application.Name, &application.Token, &application.Type, &application.Status,
		&application.Sort, &application.Level, &application.Level, &application.TenantId, &application.Created, &application.Updated); err != nil {
		glog.Error(err)

		return nil, err
	}

	return &application, nil
}
