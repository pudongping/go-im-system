package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的 channel
	Message chan string
}

// 创建一个 server 的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听 message 广播消息 channel 的 goroutine，一旦有消息就发送给全部的在线 user
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message

		// 将 msg 发送给全部的在线 user
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// 广播消息的方法
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	// 将消息发送到管道中
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	fmt.Println("当前连接成功！")

	user := NewUser(conn)

	// 用户上线，将用户加入到 onlineMap 中
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// 广播当前用户上线消息
	s.BroadCast(user, "用户已上线")

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			// 客户端合法关闭
			if n == 0 {
				s.BroadCast(user, "下线")
				return
			}

			if err != nil && err != io.EOF {
				// 有错误
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息（需要去除末尾的 '\n'）
			msg := string(buf[:n-1])

			// 将得到的消息进行广播
			s.BroadCast(user, msg)

		}
	}()

	// 当前 handler 阻塞
	select {}

}

// 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听 message 的 goroutine
	go s.ListenMessager()

	for {
		// accept 接收
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler
		go s.Handler(conn)
	}

}
