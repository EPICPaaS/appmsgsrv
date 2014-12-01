package main

import (
	"fmt"
	"net"
	"os"
	//"strconv"
	"sync"
	"time"
)

func tcpClient() {

	tcpAddr, err := net.ResolveTCPAddr("tcp", "115.29.226.14:6969")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	conn.SetKeepAlive(true)
	//链接添加成功 {userI_id}_{device_type}-{real_device_id}@{xx}
	key := "23370005043469383_Android-123@user"
	proto := []byte(fmt.Sprintf("*3\r\n$3\r\nsub\r\n$%d\r\n%s\r\n$2\r\n30\r\n", len(key), key))
	var w sync.WaitGroup
	w.Add(1)
	go readMsg(conn, w)
	if i, err := conn.Write(proto); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("成功：", i)
	}
	ht := time.NewTicker(10 * time.Second)

	for _ = range ht.C {
		if _, err := conn.Write([]byte("h")); err != nil {
			w.Done()
			fmt.Println("发送信息失败：", err)
		}
	}
	w.Wait()
	fmt.Println("结束")
}
func readMsg(conn net.Conn, w sync.WaitGroup) {
	msg := make([]byte, 512)

	for {
		n, err := conn.Read(msg)
		if err == nil {
			fmt.Println("读取到消息：", string(msg[:n]))
		} else {
			fmt.Println("读取:", err)
			w.Done()
			break
		}
	}
}
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
	}
}
