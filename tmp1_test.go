package lkws

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestAA(t *testing.T) {
	wss := NewWsServer()
	go wss.StartServer("127.0.0.1:17777","/msg")
	for{
		data1 := strconv.Itoa(rand.Intn(10000))
		wss.BroadCastMsg(data1)
		//time.Sleep(1*time.Second)
		data2 := strconv.Itoa(rand.Intn(10000))
		wss.BroadCastMsg(data2)
		fmt.Println("广播！ ---> "+data1 + "  "+data2)
		time.Sleep(time.Second*5)
	}
}