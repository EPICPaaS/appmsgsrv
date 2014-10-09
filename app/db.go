package app

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"os"
)

// 数据库操作句柄.
var MySQL *sql.DB

// 初始化数据库连接.
func InitDB() {
	glog.Info("Connecting DB....")

	var err error
	MySQL, err = sql.Open("mysql", Conf.AppDBURL)

	if nil != err {
		glog.Error(err)
		os.Exit(-1)
	}

	glog.Infof("DB max idle conns [%d]", Conf.AppDBMaxIdleConns)
	glog.Infof("DB max open conns [%d]", Conf.AppDBMaxOpenConns)

	MySQL.SetMaxIdleConns(Conf.AppDBMaxIdleConns)
	MySQL.SetMaxOpenConns(Conf.AppDBMaxOpenConns)

	glog.Info("DB connected")
}

// 关闭数据库连接.
func CloseDB() {
	MySQL.Close()
}
