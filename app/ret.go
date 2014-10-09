package app

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"time"
)

const (
	OK             = 0
	NotFoundServer = 1001
	NotFound       = 65531
	TooLong        = 65532
	AuthErr        = 65533
	ParamErr       = 65534
	InternalErr    = 65535
)

// 响应基础结构.
type baseResponse struct {
	Ret    int    `json:"ret"`
	ErrMsg string `json:"errMsg"`
}

// retWrite marshal the result and write to client(get).
func RetWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, callback string, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		glog.Errorf("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := ""
	if callback == "" {
		// Normal json
		dataStr = string(data)
	} else {
		// Jsonp
		dataStr = fmt.Sprintf("%s(%s)", callback, string(data))
	}
	if n, err := w.Write([]byte(dataStr)); err != nil {
		glog.Errorf("w.Write(\"%s\") error(%v)", dataStr, err)
	} else {
		glog.V(4).Infof("w.Write(\"%s\") write %d bytes", dataStr, n)
	}

	glog.V(4).Infof("req: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

// retPWrite marshal the result and write to client(post).
func RetPWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, body *string, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		glog.Errorf("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := string(data)
	if n, err := w.Write([]byte(dataStr)); err != nil {
		glog.Errorf("w.Write(\"%s\") error(%v)", dataStr, err)
	} else {
		glog.V(4).Infof("w.Write(\"%s\") write %d bytes", dataStr, n)
	}

	glog.V(4).Infof("req: \"%s\", post: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), *body, dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

// 带 Content-Type=application/json 头 写 JSON 数据，为了保持原有 gopush-cluster 的兼容性，所以新加了这个函数.
func RetPWriteJSON(w http.ResponseWriter, r *http.Request, res map[string]interface{}, body *string, start time.Time) {
	w.Header().Set("Content-Type", "application/json")

	RetPWrite(w, r, res, body, start)
}
