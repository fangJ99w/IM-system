package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	Conn       net.Conn
	flag       int
}

func Newclient(ServerIP string, ServerPort int) *Client {
	client := &Client{
		ServerIP:   ServerIP,
		ServerPort: ServerPort,
		flag:       999,
	}
	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ServerIP, ServerPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	client.Conn = conn
	return client
}

// 处理server回应的消息，直接显示到标准输出即可
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.Conn)
	/*
		后者一旦有数据，就会立马拷贝到前者标准输出上，永久阻塞监听
	*/
}

func (client *Client) menu() bool {
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	var flag int
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("输入合法数字")
		return false
	}
}

var ServerIp string
var ServerPort int

func (client *Client) Updatename() bool {
	fmt.Println(">>>Please input your new name.")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}
func (client *Client) PublicChat() {
	var chatmsg string
	fmt.Println("请输入聊天内容，exit 退出")
	fmt.Scanln(&chatmsg)

	for chatmsg != "exit" {
		//发给服务器

		//消息不为空则发送
		if len(chatmsg) != 0 {
			sendMsg := chatmsg + "\n"
			_, err := client.Conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}
		chatmsg = ""
		fmt.Println("请输入聊天内容，exit 退出")
		fmt.Scanln(&chatmsg)
	}
}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendmsg := "who\n"
	_, err := client.Conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}
func (client *Client) PrivateChat() {
	var remoteName string
	var chatmsg string
	client.SelectUsers()
	fmt.Println(">>>请输入聊天对象，exit 退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>请输入消息内容，exit退出： ")
		fmt.Scanln(&chatmsg)
		for chatmsg != "exit" {
			//消息不为空则发送
			if len(chatmsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatmsg + "\n\n"
				_, err := client.Conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}

			chatmsg = ""
			fmt.Println(">>>请输入消息内容，exit退出： ")
			fmt.Scanln(&chatmsg)
		}
		client.SelectUsers()
		fmt.Println(">>>请输入聊天对象，exit 退出")
		_, _ = fmt.Scanln(&remoteName)
	}
}
func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		switch client.flag {

		case 1:
			fmt.Println("公聊模式选择....")
			client.PublicChat()
			break
		case 2:
			fmt.Println("私聊模式选择....")
			client.PrivateChat()
			break
		case 3:
			fmt.Println("更新用户名选择...")
			client.Updatename()
			break

		}
		fmt.Printf("\033[2J\033[H")
	}
}

// ./cilent -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&ServerIp, "ip", "127.0.0.1", "设置默认服务器IP地址（默认是127.0.0.1）")
	flag.IntVar(&ServerPort, "port", 8888, "设置默认服务器端口（默认是8888）")

}

func main() {
	//命令行解析
	flag.Parse()
	client := Newclient(ServerIp, ServerPort)
	if client == nil {
		fmt.Println("client is nil")
		return
	}
	go client.DealResponse()
	println(">>>>>>>连接成功>>>>>>")
	//启动业务
	client.Run()
}
