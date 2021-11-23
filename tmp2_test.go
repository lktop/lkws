package lkws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"
)


func TestB(t *testing.T) {
	for {
		cliConn ,err:= NewWsClientConn("ws://127.0.0.1:17777/msg")
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

func sdsad(){
	for {
		data1 := strconv.Itoa(rand.Intn(10000))
		data2 := strconv.Itoa(rand.Intn(10000))
		hub.broadcast <- []byte(data1)
		hub.broadcast <- []byte(data2)
		fmt.Println("广播！ ---> "+data1 + "  "+data2)
		time.Sleep(2*time.Second)
	}
}


func TestAC(t *testing.T) {
	//后台启动处理器
	go hub.run()
	go sdsad()
	http.HandleFunc("/chat", wsHandle)         //将chat请求交给wshandle处理
	http.ListenAndServe("127.0.0.1:8888", nil) //开始监听
}


//定义一个websocket处理器，用于收集消息和广播消息
type Hub struct {
	//用户列表，保存所有用户
	userList map[*User]bool
	//注册chan，用户注册时添加到chan中
	register chan *User
	//注销chan，用户退出时添加到chan中，再从map中删除
	unregister chan *User
	//广播消息，将消息广播给所有连接
	broadcast chan []byte
}

//定义一个websocket连接对象，连接中包含每个连接的信息
type User struct {
	conn *websocket.Conn
	msg  chan []byte
}

//定义一个升级器，将普通的http连接升级为websocket连接
var up = &websocket.Upgrader{
	//定义读写缓冲区大小
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	//校验请求
	CheckOrigin: func(r *http.Request) bool {
		//如果不是get请求，返回错误
		if r.Method != "GET" {
			fmt.Println("请求方式错误")
			return false
		}
		//如果路径中不包括chat，返回错误
		if r.URL.Path != "/chat" {
			fmt.Println("请求路径错误")
			return false
		}
		//还可以根据其他需求定制校验规则
		return true
	},
}
//初始化处理中心，以便调用
var hub = &Hub{
	userList:   make(map[*User]bool),
	register:   make(chan *User),
	unregister: make(chan *User),
	broadcast:  make(chan []byte),
}

func wsHandle(w http.ResponseWriter, r *http.Request) {
	//通过升级后的升级器得到链接
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("获取连接失败:", err)
		return
	}
	//连接成功后注册用户
	user := &User{
		conn: conn,
		msg:  make(chan []byte),
	}
	hub.register <- user
	defer func() {
		hub.unregister <- user
	}()
	//得到连接后，就可以开始读写数据了
	go read(user)
	write(user)
}
func read(user *User) {

	//从连接中循环读取信息
	for {
		_, msg, err := user.conn.ReadMessage()
		if err != nil {
			fmt.Println("用户退出:",user.conn.RemoteAddr().String())
			hub.unregister<-user
			break
		}
		//将读取到的信息传入websocket处理器中的broadcast中，
		hub.broadcast <- msg
	}
}
func write(user *User) {
	for data := range user.msg {
		err := user.conn.WriteMessage(1, data)
		if err != nil {
			fmt.Println("写入错误")
			break
		}
	}
}

//处理中心处理获取到的信息
func (h *Hub) run() {
	for {
		select {
		//从注册chan中取数据
		case user := <-h.register:
			//取到数据后将数据添加到用户列表中
			h.userList[user] = true
		case user := <-h.unregister:
			//从注销列表中取数据，判断用户列表中是否存在这个用户，存在就删掉
			if _, ok := h.userList[user]; ok {
				delete(h.userList, user)
			}
		case data := <-h.broadcast:
			//从广播chan中取消息，然后遍历给每个用户，发送到用户的msg中
			for u := range h.userList {
				select {
				case u.msg <- data:
				}
			}
		}
	}
}

func TestTY(t *testing.T) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8888/chat", nil)
	if err != nil {
		fmt.Println("错误信息:", err)
	}
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("错误信息:", err)
			break
		}
		if err == io.EOF {
			continue
		}
		fmt.Println("获取到的信息:", string(msg))
	}
}