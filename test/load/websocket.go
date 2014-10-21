package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"code.google.com/p/go.net/websocket"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) != 4 {
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
	}

	time.Sleep(15 * time.Minute)
	fmt.Println(g)

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
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
