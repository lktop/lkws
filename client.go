package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
)



func NewWsClientConn(addr string)(*websocket.Conn,error){
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil,err
	}
	return conn,nil
}


func WsClientGetMsg(conn *websocket.Conn,revChan chan string) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("错误信息:", err)
			close(revChan)
			break
		}
		if err == io.EOF {
			continue
		}
		revChan <- string(msg)
	}
}