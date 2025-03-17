package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	Conn   net.Conn
	server *Server
}

// 创建用户API
func Newuser(conn net.Conn, server *Server) *User {
	useraddr := conn.RemoteAddr().String()

	user := &User{
		Name:   useraddr,
		Addr:   useraddr,
		C:      make(chan string),
		Conn:   conn,
		server: server,
	}
	go user.Listen()
	return user
}

// 监听当前user channel 方法。一旦有消息，就发送给对端客户端
func (this *User) Listen() {
	for {
		msg := <-this.C
		this.Conn.Write([]byte(msg + "\n"))
	}
}

// user 上线
func (this *User) Online() {

	this.server.maplock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.maplock.Unlock()
	//上线
	this.server.Broadcast(this, "已上线")
}

// user下线
func (this *User) Offline() {
	this.server.maplock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.maplock.Unlock()
	//xia线
	this.server.Broadcast(this, "已下线")
}

// 给当前user的客户端发送消息
func (this *User) Send(msg string) {
	this.Conn.Write([]byte(msg + "\n"))
}
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询在线用户有哪些
		this.server.maplock.Lock()
		for _, user := range this.server.OnlineMap {
			OnlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线。。。\n"
			this.Send(OnlineMsg)
		}
		this.server.maplock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式 rename|Zhangsan
		newName := strings.Split(msg, "|")[1]
		//判断当前newname是否被占用
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.Send("当前用户名被使用\n")

		} else {
			this.server.maplock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.maplock.Unlock()
			this.Name = newName
			this.Send(msg)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//1.获取对方用户名
		remotename := strings.Split(msg, "|")[1]
		if remotename == "" {
			this.Send("格式不正确,请使用 “tO|张三|你好”这种格式 \n")
			return
		}
		//2.得到对方user对象
		remoteuser, ok := this.server.OnlineMap[remotename]
		if !ok {
			this.Send("该用户名不存在")
			return
		}
		//3.获取消息内容，通过user对象发送消息内容。
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.Send("消息内容为空\n")
			return
		}
		remoteuser.Send(this.Name + "给你说" + content)
	} else {
		this.server.Broadcast(this, msg)
	}

}
