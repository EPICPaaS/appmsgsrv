package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// 数据库操作句柄.
var MySQL *sql.DB

// 初始化数据库连接.
func InitDB() {
	logger.Info("Connecting DB....")

	var err error
	MySQL, err = sql.Open("mysql", Conf.AppDBURL)

	if nil != err {
		logger.Error(err)
	}

	// 实际测试一次
	test := 0
	if err := MySQL.QueryRow("SELECT 1").Scan(&test); err != nil {
		logger.Error(err)
	}

	logger.Infof("DB max idle conns [%d]", Conf.AppDBMaxIdleConns)
	logger.Infof("DB max open conns [%d]", Conf.AppDBMaxOpenConns)

	MySQL.SetMaxIdleConns(Conf.AppDBMaxIdleConns)
	MySQL.SetMaxOpenConns(Conf.AppDBMaxOpenConns)

	logger.Info("DB connected")
}

// 关闭数据库连接.
func CloseDB() {
	MySQL.Close()
}
