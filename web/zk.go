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
	"github.com/EPICPaaS/gopush-cluster/app"
	myrpc "github.com/EPICPaaS/gopush-cluster/rpc"
	myzk "github.com/EPICPaaS/gopush-cluster/zk"
	"github.com/golang/glog"
	"github.com/samuel/go-zookeeper/zk"
)

func InitZK() (*zk.Conn, error) {
	conn, err := myzk.Connect(app.Conf.ZookeeperAddr, app.Conf.ZookeeperTimeout)
	if err != nil {
		glog.Errorf("zk.Connect() error(%v)", err)
		return nil, err
	}
	myrpc.InitComet(conn, app.Conf.ZookeeperCometPath, app.Conf.RPCRetry, app.Conf.RPCPing)
	myrpc.InitMessage(conn, app.Conf.ZookeeperMessagePath, app.Conf.RPCRetry, app.Conf.RPCPing)
	return conn, nil
}
