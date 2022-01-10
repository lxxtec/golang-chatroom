package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	//服务器IP
	Ip   string
	//服务器端口
	Port int
	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	//消息广播的channel
	Message chan string
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return server
}

// 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err", err)
		return
	}
	fmt.Print("socket listening...")
	// close listen socket
	defer listener.Close()

	//启动监听Message的goroutine
	go s.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept err:", err)
		}
		// do handler
		go s.Handler(conn)
	}
}

// 监听Message广播channel的goroutine，一旦有消息就发送给全体在线user
func (s *Server) ListenMessager(){
	for{
		msg := <- s.Message
		// 将msg 发送给全部在线的user
		s.mapLock.Lock()
		for _,cli:=range s.OnlineMap{
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}


// 广播消息的方法
func (s *Server) BroadCast(user *User,msg string){
	sendMsg:="["+user.Addr+"]"+user.Name+":"+msg
	s.Message <- sendMsg
}


func (s *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	//fmt.Println("连接成功！！")
	user:=NewUser(conn,s)
	// 用户上线
	user.Online()


	//监听用户是否活跃的channel
	isLive:=make(chan bool)

	//接受客户端发送的消息
	go func ()  {
		buf:=make([]byte,4096)
		for{
			n,err:=conn.Read(buf)
			if n==0{
				user.Offline()
				return
			}
			if err!=nil && err !=io.EOF {
				fmt.Println("Conn Read err",err)
				return
			}
			//提取用户的消息(去除\n)
			msg:=string(buf[:n-1])
			//用户针对msg进行处理
			user.DoMessage(msg)
			//用户的任意消息，代表当前用户是活跃的
			isLive<-true
		}
	}()

	// 当前handler阻塞
	for{
		select{
		case <-isLive:
			//当前用户是活跃的，应该重置定时器
			//这里不错任何事情，为了激活select，更新下面的定时器
		case <-time.After(time.Second*50):
			//已经超时，就当前客户端强制关闭
			user.SendMsg("你被踢了")
			//销魂资源
			close(user.C)
			//从表中剔除
			delete(s.OnlineMap,user.Name)
			//关闭连接
			conn.Close()
			//退出当前handler
			return
		}
	}
}
