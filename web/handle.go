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
	"github.com/golang/glog"
	"net/http"
	"strconv"
	"time"
)

// GetServer handle for server get
func GetServer0(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	params := r.URL.Query()
	key := params.Get("key")
	callback := params.Get("callback")
	protoStr := params.Get("proto")
	res := map[string]interface{}{"ret": app.OK, "msg": "ok"}
	defer app.RetWrite(w, r, res, callback, time.Now())
	if key == "" {
		res["ret"] = app.ParamErr
		return
	}
	proto, err := strconv.Atoi(protoStr)
	if err != nil {
		glog.Errorf("strconv.Atoi(\"%s\") error(%v)", protoStr, err)
		res["ret"] = app.ParamErr
		return
	}
	// Match a push-server with the value computed through ketama algorithm
	node := myrpc.GetComet(key)
	if node == nil {
		res["ret"] = app.NotFoundServer
		return
	}
	addr := node.Addr[proto]
	if addr == nil || len(addr) == 0 {
		res["ret"] = app.NotFoundServer
		return
	}
	server := ""
	// Select the best ip
	if app.Conf.Router != "" {
		server = routerCN.SelectBest(r.RemoteAddr, addr)
		glog.V(5).Infof("select the best ip:\"%s\" match with remoteAddr:\"%s\" , from ip list:\"%v\"", server, r.RemoteAddr, addr)
	}
	if server == "" {
		glog.V(5).Infof("remote addr: \"%s\" chose the ip: \"%s\"", r.RemoteAddr, addr[0])
		server = addr[0]
	}
	res["data"] = map[string]interface{}{"server": server}
	return
}

// GetOfflineMsg get offline mesage http handler.
func GetOfflineMsg0(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	params := r.URL.Query()
	key := params.Get("key")
	midStr := params.Get("mid")
	callback := params.Get("callback")
	res := map[string]interface{}{"ret": app.OK, "msg": "ok"}
	defer app.RetWrite(w, r, res, callback, time.Now())
	if key == "" || midStr == "" {
		res["ret"] = app.ParamErr
		return
	}
	mid, err := strconv.ParseInt(midStr, 10, 64)
	if err != nil {
		res["ret"] = app.ParamErr
		glog.Errorf("strconv.ParseInt(\"%s\", 10, 64) error(%v)", midStr, err)
		return
	}
	// RPC get offline messages
	reply := &myrpc.MessageGetResp{}
	args := &myrpc.MessageGetPrivateArgs{MsgId: mid, Key: key}
	client := myrpc.MessageRPC.Get()
	if client == nil {
		res["ret"] = app.InternalErr
		return
	}
	if err := client.Call(myrpc.MessageServiceGetPrivate, args, reply); err != nil {
		glog.Errorf("myrpc.MessageRPC.Call(\"%s\", \"%v\", reply) error(%v)", myrpc.MessageServiceGetPrivate, args, err)
		res["ret"] = app.InternalErr
		return
	}
	omsgs := []string{}
	opmsgs := []string{}
	for _, msg := range reply.Msgs {
		omsg, err := msg.OldBytes()
		if err != nil {
			res["ret"] = app.InternalErr
			return
		}
		omsgs = append(omsgs, string(omsg))
	}

	if len(omsgs) == 0 {
		return
	}

	res["data"] = map[string]interface{}{"msgs": omsgs, "pmsgs": opmsgs}
	return
}

// GetTime get server time http handler.
func GetTime0(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	params := r.URL.Query()
	callback := params.Get("callback")
	res := map[string]interface{}{"ret": app.OK, "msg": "ok"}
	now := time.Now()
	defer app.RetWrite(w, r, res, callback, now)
	res["data"] = map[string]interface{}{"timeid": now.UnixNano() / 100}
	return
}
