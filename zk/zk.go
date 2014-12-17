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

// github.com/samuel/go-zookeeper
// Copyright (c) 2013, Samuel Stauffer <samuel@descolada.com>
// All rights reserved.

package zk

import (
	"errors"
	"github.com/b3log/wide/log"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
)

var (
	// error
	ErrNoChild      = errors.New("zk: children is nil")
	ErrNodeNotExist = errors.New("zk: node not exist")
	logger          = log.NewLogger(os.Stdout)
)

// Connect connect to zookeeper, and start a goroutine log the event.
func Connect(addr []string, timeout time.Duration) (*zk.Conn, error) {
	conn, session, err := zk.Connect(addr, timeout)
	if err != nil {
		logger.Errorf("zk.Connect(\"%v\", %d) error(%v)", addr, timeout, err)
		return nil, err
	}
	go func() {
		for {
			event := <-session
			logger.Tracef("zookeeper get a event: %s", event.State.String())
		}
	}()
	return conn, nil
}

// Create create zookeeper path, if path exists ignore error
func Create(conn *zk.Conn, fpath string) error {
	// create zk root path
	tpath := ""
	for _, str := range strings.Split(fpath, "/")[1:] {
		tpath = path.Join(tpath, "/", str)
		logger.Tracef("create zookeeper path: \"%s\"", tpath)
		_, err := conn.Create(tpath, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				logger.Warnf("zk.create(\"%s\") exists", tpath)
			} else {
				logger.Errorf("zk.create(\"%s\") error(%v)", tpath, err)
				return err
			}
		}
	}
	return nil
}

// RegisterTmp create a ephemeral node, and watch it, if node droped then send a SIGQUIT to self.
func RegisterTemp(conn *zk.Conn, fpath, data string) error {
	tpath, err := conn.Create(path.Join(fpath)+"/", []byte(data), zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		logger.Errorf("conn.Create(\"%s\", \"%s\", zk.FlagEphemeral|zk.FlagSequence) error(%v)", fpath, data, err)
		return err
	}
	logger.Tracef("create a zookeeper node:%s", tpath)
	// watch self
	go func() {
		for {
			logger.Infof("zk path: \"%s\" set a watch", tpath)
			exist, _, watch, err := conn.ExistsW(tpath)
			if err != nil {
				logger.Errorf("zk.ExistsW(\"%s\") error(%v)", tpath, err)
				logger.Warnf("zk path: \"%s\" set watch failed, kill itself", tpath)
				killSelf()
				return
			}
			if !exist {
				logger.Warnf("zk path: \"%s\" not exist, kill itself", tpath)
				killSelf()
				return
			}
			event := <-watch
			logger.Infof("zk path: \"%s\" receive a event %v", tpath, event)
		}
	}()
	return nil
}

// GetNodesW get all child from zk path with a watch.
func GetNodesW(conn *zk.Conn, path string) ([]string, <-chan zk.Event, error) {
	nodes, stat, watch, err := conn.ChildrenW(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, nil, ErrNodeNotExist
		}
		logger.Errorf("zk.ChildrenW(\"%s\") error(%v)", path, err)
		return nil, nil, err
	}
	if stat == nil {
		return nil, nil, ErrNodeNotExist
	}
	if len(nodes) == 0 {
		return nil, nil, ErrNoChild
	}
	return nodes, watch, nil
}

// GetNodes get all child from zk path.
func GetNodes(conn *zk.Conn, path string) ([]string, error) {
	nodes, stat, err := conn.Children(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, ErrNodeNotExist
		}
		logger.Errorf("zk.Children(\"%s\") error(%v)", path, err)
		return nil, err
	}
	if stat == nil {
		return nil, ErrNodeNotExist
	}
	if len(nodes) == 0 {
		return nil, ErrNoChild
	}
	return nodes, nil
}

// killSelf send a SIGQUIT to self.
func killSelf() {
	if err := syscall.Kill(os.Getpid(), syscall.SIGQUIT); err != nil {
		logger.Errorf("syscall.Kill(%d, SIGQUIT) error(%v)", os.Getpid(), err)
	}
}
