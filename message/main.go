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
	"github.com/EPICPaaS/appmsgsrv/perf"
	"github.com/EPICPaaS/appmsgsrv/process"
	"github.com/EPICPaaS/appmsgsrv/ver"
	"github.com/b3log/wide/log"
	"os"
	"runtime"
	"time"
)

var logger = log.NewLogger(os.Stdout)

func main() {
	logLevel := flag.String("log_level", "info", "logger level")
	flag.Parse()
	log.SetLevel(*logLevel)
	logger.Infof("message ver: \"%s\" start", ver.Version)

	if err := InitConfig(); err != nil {
		logger.Errorf("InitConfig() error(%v)", err)
		return
	}
	// Set max routine
	runtime.GOMAXPROCS(Conf.MaxProc)
	// start pprof http
	perf.Init(Conf.PprofBind)
	// Initialize redis
	if err := InitStorage(); err != nil {
		logger.Errorf("InitStorage() error(%v)", err)
		return
	}
	// init rpc service
	InitRPC()
	// init zookeeper
	zk, err := InitZK()
	if err != nil {
		logger.Errorf("InitZK() error(%v)", err)
		if zk != nil {
			zk.Close()
		}
		return
	}
	// if process exit, close zk
	defer zk.Close()
	// init process
	// sleep one second, let the listen start
	time.Sleep(time.Second)
	if err = process.Init(Conf.User, Conf.Dir, Conf.PidFile); err != nil {
		logger.Errorf("process.Init(\"%s\", \"%s\", \"%s\") error(%v)", Conf.User, Conf.Dir, Conf.PidFile, err)
		return
	}
	// init signals, block wait signals
	sig := InitSignal()
	HandleSignal(sig)
	// exit
	logger.Info("message stop")
}
