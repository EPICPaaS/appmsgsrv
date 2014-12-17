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

package db

import (
	"flag"
	"fmt"
	"github.com/Terry-Mao/goconf"
	"github.com/b3log/wide/log"
	"os"
)

var (
	Conf     *Config
	confFile string
)

var logger = log.NewLogger(os.Stdout)

// InitConfig initialize config file path
func init() {
	flag.StringVar(&confFile, "dbc", "./db.conf", " set db config file path")
}

type Config struct {
	AppDBURL          string `goconf:"base:app.dbURL"`
	AppDBMaxIdleConns int    `goconf:"base:app.dbMaxIdleConns"`
	AppDBMaxOpenConns int    `goconf:"base:app.dbMaxOpenConns"`
}

// InitConfig init configuration file.
func InitConfig() error {

	gconf := goconf.New()
	if err := gconf.Parse(confFile); err != nil {
		fmt.Println("confFile", confFile)
		logger.Errorf("goconf.Parse(\"%s\") error(%v)", confFile, err)
		return err
	}
	Conf = &Config{}
	if err := gconf.Unmarshal(Conf); err != nil {
		logger.Errorf("goconf.Unmarshall() error(%v)", err)
		return err
	}

	return nil
}
