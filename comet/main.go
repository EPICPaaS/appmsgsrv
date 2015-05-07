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
	"github.com/EPICPaaS/appmsgsrv/db"
	"github.com/EPICPaaS/appmsgsrv/perf"
	"github.com/EPICPaaS/appmsgsrv/process"
	"github.com/EPICPaaS/appmsgsrv/session"
	"github.com/EPICPaaS/appmsgsrv/ver"
	"github.com/b3log/wide/log"
	"os"
	"runtime"
	"time"
)

var logger = log.NewLogger(os.Stdout)

func main() {

	logLevel := flag.String("log_level", "info", "logger level")
	// parse cmd-line arguments
	flag.Parse()
	log.SetLevel(*logLevel)
	logger.Infof("comet ver: \"%s\" start", ver.Version)

	// init config
	if err := InitConfig(); err != nil {
		logger.Errorf("InitConfig() error(%v)", err)
		return
	}
	//init db config
	if err := db.InitConfig(); err != nil {
		logger.Error("db-InitConfig() wrror(%v)", err)
		return
	}

	db.InitDB()
	defer db.CloseDB()

	// set max routine
	runtime.GOMAXPROCS(Conf.MaxProc)
	// start pprof
	perf.Init(Conf.PprofBind)
	// create channel
	// if process exit, close channel
	UserChannel = NewChannelList()
	defer UserChannel.Close()
	// start stats
	StartStats()
	// start rpc
	StartRPC()
	// start comet
	StartComet()
	// init zookeeper
	zkConn, err := InitZK()
	if err != nil {
		logger.Errorf("InitZookeeper() error(%v)", err)
		return
	}
	// if process exit, close zk
	defer zkConn.Close()
	// init process
	// sleep one second, let the listen start
	time.Sleep(time.Second)
	if err = process.Init(Conf.User, Conf.Dir, Conf.PidFile); err != nil {
		logger.Errorf("process.Init(\"%s\", \"%s\", \"%s\") error(%v)", Conf.User, Conf.Dir, Conf.PidFile, err)
		return
	}
	go session.ScanSession()
	// init signals, block wait signals
	signalCH := InitSignal()
	HandleSignal(signalCH)
	// exit
	logger.Info("comet stop")
}
