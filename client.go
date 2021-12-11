package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string   // 链接 ip 地址
	ServerPort int      // 链接端口
	Name       string   // 用户名
	conn       net.Conn // 链接句柄
	flag       int      // 当前 client 的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 链接 server
	// 创建一个会话
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	// 返回对象
	return client
}

// 处理 server 回应的消息，直接显示到标准输出即可
func (c *Client) DealResponse() {
	// 一旦 c.conn 有数据，就直接 copy 到 stdout 标准输出上，永久阻塞监听
	io.Copy(os.Stdout, c.conn)

	//for {
	//	buf := make()
	//	c.conn.Read(buf)
	//	fmt.Println(buf)
	//}
}

func (c *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	// 等待用户键盘输入
	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}

}

func (c *Client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容，exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送给服务器

		// 消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)

	}

}

func (c *Client) UpdateName() bool {

	fmt.Println(">>>请输入用户名：")
	fmt.Scanln(&c.Name)

	sendMsg := fmt.Sprintf("rename|%s\n", c.Name)
	//sendMsg := "rename|" + c.Name + "\n"
	// 发送消息
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (c *Client) Run() {
	if c.flag != 0 {
		for c.menu() != true {

		}

		// 根据 flag 不同的模式处理不同的业务
		switch c.flag {
		case 1:
			// 公聊模式
			//fmt.Println("公聊模式选择")
			c.PublicChat()
			break
		case 2:
			// 私聊模式
			fmt.Println("私聊模式选择")
			break
		case 3:
			// 更新用户名
			//fmt.Println("更新用户名选择")
			c.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
// 第 4 个参数为，当执行 ./client -h 时的提示信息
// init 函数会在 main 函数之前被自动调用
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器 IP 地址（默认是 127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认是 8888 ）")
}

func main() {

	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>> 链接服务器失败……")
		return
	}

	// 单独开启一个 goroutine 去处理 server 的回执消息
	go client.DealResponse()

	fmt.Println(">>>>>>>> 链接服务器成功……")

	// 启动客户端的业务
	client.Run()
}
