package lkws

import (
	"fmt"
	"testing"
	"time"
)


func TestB(t *testing.T) {
	for {
		cliConn ,err:= NewWsClientConn("ws://127.0.0.1:18888/msg")
		if err!=nil{
			fmt.Println("连接失败！  -->  "+ err.Error())
			time.Sleep(3*time.Second)
			fmt.Println("尝试重连！")
			continue
		}
		fmt.Println("连接成功  ---> local:"+cliConn.LocalAddr().String())
		revChan := make(chan string)
		go WsClientGetMsg(cliConn,revChan)
		for data := range revChan{
			fmt.Println(data)
		}
		fmt.Println("连接断开！")
		time.Sleep(3*time.Second)
		fmt.Println("尝试重连！")
	}
}

