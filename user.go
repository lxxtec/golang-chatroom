package main

import (
	"net"
	"strings"
)

type User struct{
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建用户api
func NewUser(conn net.Conn,server *Server) *User{
	userAddr:=conn.RemoteAddr().String()
 
	user:=&User{
		Name: userAddr,
		Addr: userAddr,
		C:	make(chan string),
		conn: conn,
		server: server,
	}
	// 每创建一个用户都开启一个协程
	go user.ListenMessage()
	return user
}

// 监听当前User channel的方法，一旦有消息，就直接发送个客户端
func (s *User) ListenMessage(){
	for{
		msg := <- s.C
		s.conn.Write([]byte(msg+"\n"))
	}
}
//用户上线业务
func (s *User) Online(){
	// 用户上线，将用户加入到onlinemap中
	s.server.mapLock.Lock()
	s.server.OnlineMap[s.Name]=s
	s.server.mapLock.Unlock()

	// 广播当前用户上线消息
	s.server.BroadCast(s,"已上线")
}
//用户下线业务
func (s *User) Offline(){
	// 用户下线，将用户从onlinemap中删除
	s.server.mapLock.Lock()
	delete(s.server.OnlineMap,s.Name)
	s.server.mapLock.Unlock()

	// 广播当前用户上线消息
	s.server.BroadCast(s,"已下线")
}

//给当前user对应的客户端发送消息
func (s *User) SendMsg(msg string){
	s.conn.Write([]byte(msg))
}

//用户处理消息
func (s *User) DoMessage(msg string){
	if msg=="who"{
		//查询当前在线用户都有哪些
		s.server.mapLock.Lock()
		for _,user:=range s.server.OnlineMap{
			onlineMsg:="["+user.Addr+"]"+user.Name+":"+ "在线...\n"
			s.SendMsg(onlineMsg)
		}
		s.server.mapLock.Unlock()
	} else if len(msg)>7 && msg[:7]=="rename|"{
		//消息格式：rename|张三
		newName:=strings.Split(msg,"|")[1]
		s.ChangeName(newName)
	} else if len(msg)>4 && msg[:3]=="to|"{
		//消息格式：to|张三|消息内容
		//1. 获取对方用户名
		remoteName:=strings.Split(msg,"|")[1]
		if remoteName == ""{
			s.SendMsg("消息格式不正确\n")
			return
		}
		//2. 根据用户名得到对方user对象
		remoteUser,ok:=s.server.OnlineMap[remoteName]
		if !ok{
			s.SendMsg("该用户名不存在")
		}
		//3. 获取消息内容，通过对方User对象将消息内容发送过去
		content:=strings.Split(msg,"|")[2]
		if content==""{
			s.SendMsg("无内容，请重试！")
			return
		}
		remoteUser.SendMsg(s.Name+"对您说:"+content+"\n")
	}


	//s.server.BroadCast(s,msg)
}

// 修改用户名
func (s *User) ChangeName(name string){
	//判断name 是否存在
	_,ok:=s.server.OnlineMap[name]
	if ok{
		s.SendMsg("当前用户名已被使用\n")
	}else {
		s.server.mapLock.Lock()
		delete(s.server.OnlineMap,s.Name)
		s.server.OnlineMap[name]=s
		s.server.mapLock.Unlock()
		s.Name=name
		s.SendMsg("您已更新用户名为"+s.Name+"\n")
	}
}
