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
	"runtime"
	"time"

	"github.com/EPICPaaS/appmsgsrv/app"
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/appmsgsrv/perf"
	"github.com/EPICPaaS/appmsgsrv/process"
	"github.com/EPICPaaS/appmsgsrv/ver"
	"github.com/golang/glog"
)

func main() {
	var err error
	// Parse cmd-line arguments
	flag.Parse()
	glog.Infof("web ver: \"%s\" start", ver.Version)
	defer glog.Flush()
	if err = app.InitConfig(); err != nil {
		glog.Errorf("InitConfig() error(%v)", err)
		return
	}
	//init db config
	if err = db.InitConfig(); err != nil {
		glog.Error("db-InitConfig() error(%v)", err)
		return
	}
	// Set max routine
	runtime.GOMAXPROCS(app.Conf.MaxProc)
	// init zookeeper
	zkConn, err := InitZK()
	if err != nil {
		glog.Errorf("InitZookeeper() error(%v)", err)
		return
	}
	// if process exit, close zk
	defer zkConn.Close()
	// start pprof http
	perf.Init(app.Conf.PprofBind)
	// Init network router
	if app.Conf.Router != "" {
		if err := InitRouter(); err != nil {
			glog.Errorf("InitRouter() failed(%v)", err)
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
		glog.Errorf("process.Init() error(%v)", err)
		return
	}

	defer db.MySQL.Close()

	//初始化配额配置
	app.InitQuotaAll()
	go app.LoadQuotaAll()

	// init signals, block wait signals
	signalCH := InitSignal()
	HandleSignal(signalCH)

	glog.Info("web stop")
}
