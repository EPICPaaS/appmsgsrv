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

package app

import (
	"flag"
	"fmt"
	"github.com/Terry-Mao/goconf"
	"runtime"
	"time"
)

var (
	Conf     *Config
	confFile string
)

const (
	/*后缀信息*/
	TENANT_SUFFIX = "@tenant"
	ORG_SUFFIX    = "@org"
	QUN_SUFFIX    = "@qun"
	USER_SUFFIX   = "@user"
	APP_SUFFIX    = "@app"
)

// InitConfig initialize config file path
func init() {
	flag.StringVar(&confFile, "c", "./web.conf", " set web config file path")
}

type Config struct {
	HttpBind             []string          `goconf:"base:http.bind:,"`
	AdminBind            []string          `goconf:"base:admin.bind:,"`
	AppBind              []string          `goconf:"base:app.bind:,"`
	MaxProc              int               `goconf:"base:maxproc"`
	PprofBind            []string          `goconf:"base:pprof.bind:,"`
	User                 string            `goconf:"base:user"`
	PidFile              string            `goconf:"base:pidfile"`
	Dir                  string            `goconf:"base:dir"`
	Router               string            `goconf:"base:router"`
	QQWryPath            string            `goconf:"res:qqwry.path"`
	ZookeeperAddr        []string          `goconf:"zookeeper:addr:,"`
	ZookeeperTimeout     time.Duration     `goconf:"zookeeper:timeout:time"`
	ZookeeperCometPath   string            `goconf:"zookeeper:comet.path"`
	ZookeeperMessagePath string            `goconf:"zookeeper:message.path"`
	RPCRetry             time.Duration     `goconf:"rpc:retry:time"`
	RPCPing              time.Duration     `goconf:"rpc:ping:time"`
	RedisSource          map[string]string `goconf:"-"`
	RedisIdleTimeout     time.Duration     `goconf:"redis:timeout:time"`
	RedisMaxIdle         int               `goconf:"redis:idle"`
	RedisMaxActive       int               `goconf:"redis:active"`
	RedisMaxStore        int               `goconf:"redis:store"`
	RedisKetamaBase      int               `goconf:"redis:ketama.base"`
	TokenExpire          int               `goconf:"token:expire"`
	ApnsType             string            `goconf:"apns:type"`
	WeedfsAddr           string            `goconf:"weedfs:address:"`
}

// InitConfig init configuration file.
func InitConfig() error {
	gconf := goconf.New()
	if err := gconf.Parse(confFile); err != nil {
		logger.Errorf("goconf.Parse(\"%s\") error(%v)", confFile, err)
		return err
	}
	// Default config
	Conf = &Config{
		HttpBind:             []string{"localhost:80"},
		AdminBind:            []string{"localhost:81"},
		MaxProc:              runtime.NumCPU(),
		PprofBind:            []string{"localhost:8190"},
		User:                 "nobody nobody",
		PidFile:              "/tmp/gopush-cluster-web.pid",
		Dir:                  "./",
		Router:               "",
		QQWryPath:            "/tmp/QQWry.dat",
		ZookeeperAddr:        []string{":2181"},
		ZookeeperTimeout:     30 * time.Second,
		ZookeeperCometPath:   "/gopush-cluster-comet",
		ZookeeperMessagePath: "/gopush-cluster-message",
		RPCRetry:             3 * time.Second,
		RPCPing:              1 * time.Second,
		RedisSource:          make(map[string]string),
	}

	if err := gconf.Unmarshal(Conf); err != nil {
		logger.Errorf("goconf.Unmarshall() error(%v)", err)
		return err
	}

	redisAddrsSec := gconf.Get("redis.source")
	if redisAddrsSec != nil {
		for _, key := range redisAddrsSec.Keys() {
			addr, err := redisAddrsSec.String(key)
			if err != nil {
				return fmt.Errorf("config section: \"redis.addrs\" key: \"%s\" error(%v)", key, err)
			}
			Conf.RedisSource[key] = addr
		}
	}
	return nil
}
