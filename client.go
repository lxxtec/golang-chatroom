package main

import (
	"fmt"
	"net"
)

type Client struct{
	ServerIP 	string
	ServerPort 	int 
	Name 		string
	conn 		net.Conn
}

func NewClient(serverIp string,serverPort int) *Client{
	//创建客户端对象，创建连接，返回对象
	client:=&Client{
		ServerIP: serverIp,
		ServerPort: serverPort,
	}
	//连接server
	conn,err:=net.Dial("tcp",fmt.Sprintf("%s:%d",serverIp,serverPort))
	if err!=nil{
		fmt.Println("net.Dial err",err)
		return nil
	}
	client.conn=conn 
	return client
}

func main(){
	client:=NewClient("127.0.0.1",8888)
	if client==nil{
		fmt.Println(">>>>>连接服务器失败")
	}
	fmt.Println(">>>>>连接服务器成功")
	//启动客户端服务
	select{}
}