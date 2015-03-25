// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gopush-cluster.

// gopush-cluster is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gopush-cluster is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gopush-cluster.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"os"
	"runtime"
	"time"

	"github.com/EPICPaaS/appmsgsrv/app"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/appmsgsrv/perf"
	"github.com/EPICPaaS/appmsgsrv/process"
	"github.com/EPICPaaS/appmsgsrv/ver"

	"github.com/b3log/wide/log"
)

var logger = log.NewLogger(os.Stdout)

func main() {
	var err error
	// Parse cmd-line arguments
	logLevel := flag.String("log_level", "info", "logger level")
	flag.Parse()
	log.SetLevel(*logLevel)

	logger.Infof("web ver: \"%s\" start", ver.Version)

	if err = app.InitConfig(); err != nil {
		logger.Errorf("InitConfig() error(%v)", err)
		return
	}
	//init db config
	if err = db.InitConfig(); err != nil {
		logger.Error("db-InitConfig() error(%v)", err)
		return
	}
	// Set max routine
	runtime.GOMAXPROCS(app.Conf.MaxProc)
	// init zookeeper
	zkConn, err := InitZK()
	if err != nil {
		logger.Errorf("InitZookeeper() error(%v)", err)
		return
	}
	// if process exit, close zk
	defer zkConn.Close()
	// start pprof http
	perf.Init(app.Conf.PprofBind)
	// Init network router
	if app.Conf.Router != "" {
		if err := InitRouter(); err != nil {
			logger.Errorf("InitRouter() failed(%v)", err)
			return
		}
	}

	db.InitDB()
	defer db.CloseDB()

	app.InitRedisStorage()

	// start http listen.
	StartHTTP()
	// init process
	// sleep one second, let the listen start
	time.Sleep(time.Second)
	if err = process.Init(app.Conf.User, app.Conf.Dir, app.Conf.PidFile); err != nil {
		logger.Errorf("process.Init() error(%v)", err)
		return
	}

	defer db.MySQL.Close()

	//初始化配额配置
	app.InitQuotaAll()
	go app.LoadQuotaAll()

	//启动扫描过期文件
	go app.ScanExpireFileLink()

	// init signals, block wait signals
	signalCH := InitSignal()
	HandleSignal(signalCH)

	logger.Info("web stop")
}
