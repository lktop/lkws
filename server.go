package lkws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

// WsClient 定义一个websocket连接对象，连接中包含每个连接的信息
type WsClient struct {
	conn *websocket.Conn
	msg  chan string
}


// WsServer 定义一个websocket处理器，用于收集消息和广播消息
type WsServer struct {
	//连接客户端列表
	clientList map[*WsClient]bool
	//注册chan，用户注册时添加到chan中
	register chan *WsClient
	//注销chan，用户退出时添加到chan中，再从map中删除
	unregister chan *WsClient
	//广播消息，将消息广播给所有连接
	broadcast chan string
}


func (ws *WsServer)BroadCastMsg(data string){
	ws.broadcast <- data
}

func (ws *WsServer) run() {
	for {
		select {
		//从注册chan中取数据
		case cli := <-ws.register:
			//取到数据后将数据添加到客户端列表中
			ws.clientList[cli] = true
		case cli := <-ws.unregister:
			//从注销列表中取数据，判断客户端列表中是否存在这个用户，存在就删掉
			if _, ok := ws.clientList[cli]; ok {
				delete(ws.clientList, cli)
			}
		case data := <-ws.broadcast:
			//从广播chan中取消息，然后遍历给每个用户，发送到用户的msg中
			for cli := range ws.clientList {
				select {
				case cli.msg <- data:
				default:
					fmt.Println("移除客户端:",cli.conn.RemoteAddr().String())
					delete(ws.clientList, cli)
					close(cli.msg)
				}
			}
		}
	}
}


func (cli *WsClient)read(server *WsServer) {
	//从连接中循环读取信息
	for {
		_, msg, err := cli.conn.ReadMessage()
		if err != nil {
			fmt.Println("客户端下线:",cli.conn.RemoteAddr().String())
			server.unregister<-cli
			break
		}
		fmt.Println("Read Msg:"+string(msg))
	}
}


func (cli *WsClient)write() {
	for data := range cli.msg {
		err := cli.conn.WriteMessage(1, []byte(data))
		if err != nil {
			fmt.Println("写入错误")
			break
		}
	}
}


func NewWsServer()*WsServer{
	return &WsServer{
		clientList: make(map[*WsClient]bool),
		register:   make(chan *WsClient),
		unregister: make(chan *WsClient),
		broadcast:  make(chan string),
	}
}


func (ws *WsServer)StartServer(addr string,path string){
	//后台启动处理器
	go ws.run()

	up := &websocket.Upgrader{
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
			if r.URL.Path != path {
				fmt.Println("请求路径错误")
				return false
			}
			//还可以根据其他需求定制校验规则
			return true
		},
	}

	http.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		//通过升级后的升级器得到链接
		conn, err := up.Upgrade(writer, request, nil)
		if err != nil {
			fmt.Println("获取连接失败:", err)
			return
		}
		//连接成功后注册用户
		client := &WsClient{
			conn: conn,
			msg:  make(chan string,50),
		}
		ws.register <- client
		fmt.Println(conn.RemoteAddr().String()+" ---> 上线！")
		defer func() {
			ws.unregister <- client
		}()
		//得到连接后，就可以开始读写数据了
		go client.read(ws)
		client.write()
	})
	err := http.ListenAndServe(addr, nil)  //开始监听
	if err != nil {
		return
	}
}