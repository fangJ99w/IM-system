package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	maplock   sync.RWMutex
	//消息广播的channel
	Message chan string
}

// create server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听MEssage广播消息channel 的goroutine，一旦有消息就发送给所有的user
func (this *Server) Listenme() {
	for {
		msg := <-this.Message
		//将msg发送给所有的user
		this.maplock.Lock()
		for _, cil := range this.OnlineMap {
			cil.C <- msg
		}
		this.maplock.Unlock()

	}
}
func (this *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}
func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("链接成功")
	//用户上线
	user := Newuser(conn, this)

	user.Online()
	//监听用户是否活跃
	islive := make(chan bool)
	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read err :", err)
			}
			//提取用户的消息 （去除\n）
			msg := string(buf[:n-1])
			user.DoMessage(msg)

			//任意操作都会使其活跃
			islive <- true
		}

	}()

	//当前handler阻塞
	for {
		select {
		case <-islive:
		case <-time.After(time.Second * 30):
			//用户超时，将当前用户踢出
			user.Send("超时被踢了")
			//销毁用的资源
			close(user.C)
			//关闭连接
			conn.Close()
			//exit
			return // runtime.Goexit()
		}

	}
}

//open server

func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", this.Ip+":"+strconv.Itoa(this.Port))
	if err != nil {
		fmt.Println("net listen error:", err)
	}
	defer listener.Close()

	go this.Listenme()
	//accept
	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("net accept error:", err)
			continue
		}

		// do handle
		go this.Handler(conn)
	}
}
