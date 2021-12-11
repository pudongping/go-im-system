package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string // 链接 ip 地址
	ServerPort int    // 链接端口
	Name       string
	conn       net.Conn // 链接句柄
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
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

	fmt.Println(">>>>>>>> 链接服务器成功……")

	// 启动客户端的业务
	select {}
}
