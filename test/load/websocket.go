package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

//登陆消息体
var loginMsgBody = []byte(`
		{
			"baseRequest" : {
		         "token": "eflow_token"
			},
		    "sessions": ["all"],
		    "content": "Test!",
		    "msgType": "1",
			"toUserNames" : ["23622391649384525@user", "22622391649384527@user"],
		    "objectContent": {
		        "appId": "23622391649370202",
		        "appSendId": "xxxxx"
		    }
		}
	`)

//应用发送消息给服务端消息
var appSendMsgBody = []byte(`
		{
			"baseRequest" : {
		         "token": "eflow_token"
			},
		    "sessions": ["all"],
		    "content": "Test!",
		    "msgType": "1",
			"toUserNames" : ["23622391649384525@user", "22622391649384527@user"],
		    "objectContent": {
		        "appId": "23622391649370202",
		        "appSendId": "xxxxx"
		    }
		}
	`)

//设备发送消息体
var diveceMsgBody = []byte(`
		{
			"baseRequest": {
				"uid": "23622391649384525",
				"deviceID": "e907195984764735",
		        "deviceType":"iOS",
		        "token": "23622391649370004_4943f8c5-54df-4512-926d-5fb9166ba276"
			},
		    "sessions": ["all"],
			"msg": {
				"fromUserName": "23622391649384525@user",
				"toUserName": "23622391649370202@user",
				"msgType": 1,
				"content": "Test!",
		        "clientMsgId": 1407734409242
			}
		}
	`)
var errNum int = 0
var senNum int = 0

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//websocket 压测
	/*if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "请输入连接数和压测项,列如 conn/newRemove 500 11111(user_id)", os.Args[0])
		os.Exit(1)
	}
	flag := os.Args[1]
	count, err := strconv.Atoi(os.Args[2])
	checkErr(err)
	userId := os.Args[3]

	fmt.Println(flag, count)
	g := 0
	if flag == "conn" {
		fmt.Print("----开始压测在线会话，并发数为：", count)
		for i := 0; i < count; i++ {
			go starWebsocket(i, userId)
		}
	} else if flag == "newRemove" {
		for i := 0; i < count; i++ {
			go NewRemoveConn(i, userId)
		}
	}*/

	for i := 0; i < 1000; i++ {
		go sendMsg("http://10.180.120.63:8093/app/client/app/user/push", appSendMsgBody)
	}
	//sendMsg("http://10.180.120.63:8093/app/client/app/user/push", appSendMsgBody)
	time.Sleep(5 * time.Minute)
	fmt.Printf("异常次数:[%d] \n", errNum)
	fmt.Printf("总请求数:[%d]", senNum)

}

func starWebsocket(pid int, userId string) {

	origin := "http://localhost/"
	url := "ws://10.180.120.63:6968/sub?key=" + userId + "_Netscape-5-" + strconv.Itoa(pid) + "@user&heartbeat=60"

	ws, err := websocket.Dial(url, "", origin)
	checkErr(err)

	msg := make([]byte, 512)
	_, err = ws.Read(msg)
	checkErr(err)

	data := []byte("h")
	ticker := time.NewTicker(30 * time.Second)
	for _ = range ticker.C {
		websocket.Message.Send(ws, string(data))
		_, err := ws.Read(msg)
		checkErr(err)
	}

}

func NewRemoveConn(pid int, userId string) {

	for {
		count := rand.Intn(5) + 1
		origin := "http://localhost/"
		url := "ws://10.180.120.63:6968/sub?key=" + userId + "_Netscape-5-" + strconv.Itoa(pid) + ":" + strconv.Itoa(count) + "@user&heartbeat=60"
		ws, err := websocket.Dial(url, "", origin)
		checkErr(err)

		msg := make([]byte, 512)
		_, err = ws.Read(msg)
		checkErr(err)

		data := []byte("h")
		ticker := time.NewTicker(5 * time.Second)
		i := 0
		for _ = range ticker.C {
			websocket.Message.Send(ws, string(data))
			_, err := ws.Read(msg)
			checkErr(err)
			if i > count {
				ws.Close()

				break
			}

			i++
		}
	}

}

//发送post请求
func sendMsg(url string, msgBody []byte) {
	//不断发送消息
	for {

		body := bytes.NewReader(msgBody)
		res, err := http.Post(url, "text/plain;charset=UTF-8", body)
		if err != nil {
			log.Fatal(err)
			errNum++
			continue
		}
		if res.StatusCode == 200 { //请求成功
			defer res.Body.Close()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatalf("读取response信息异常：[s%]", err)
				errNum++
				continue
			}
			var args map[string]interface{}
			if err = json.Unmarshal(resBody, &args); err != nil {
				log.Fatalf("读取response信息转换为异常：[%s]", err)
				errNum++
				continue
			}
			baseResponse := args["baseResponse"].(map[string]interface{})
			ret := baseResponse["ret"].(float64)
			if ret != 0 {
				fmt.Printf("服务器处理异常，返回吗：[%d]", int(ret))
				errNum++
				continue
			}
			senNum++

		} else {
			fmt.Printf("请求失败,错误码:[%d]", res.StatusCode)
		}

	}

}
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
