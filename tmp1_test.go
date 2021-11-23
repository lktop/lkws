package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestAA(t *testing.T) {
	wss := NewWsServer()
	go wss.StartServer("127.0.0.1:18888","/msg")
	for{
		data := strconv.Itoa(rand.Intn(10000))
		wss.BroadCastMsg(data)
		fmt.Println("广播！ ---> "+data)
		time.Sleep(time.Second*5)
	}
}