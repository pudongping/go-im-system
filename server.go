package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

	user := NewUser(conn, s)

	// 用户上线操作
	user.Online()

	// 监听用户是否为活跃的 channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			// 客户端合法关闭
			if n == 0 {
				// 用户下线
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				// 有错误
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息（需要去除末尾的 '\n'）
			msg := string(buf[:n-1])

			// 用户针对 msg 进行消息处理
			user.DoMessage(msg)

			// 用户如果发送了任意消息，则代表当前用户是一个活跃中的用户
			isLive <- true

		}
	}()

	for {
		// 当前 handler 阻塞
		select {
		// 这里的 case <- isLive 必须要写在上面，因为一旦触发了这里的 case，那么其实下面的 case 也会尝试着去触发
		// 在第一个 case 中没有写 break 或者 return 语句的情况下
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器

			// 不做任何事情，是为了激活 select 语句，更新下面的定时器

		// 启动一个定时器，这个定时器其实是一个 channel
		// 10s 后会被超时
		case <-time.After(time.Second * 10):
			// 已经超时
			// 将当前的 user 强制关闭

			user.SendMsg("你被踢掉了\n")

			// 销毁用户的资源
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出当前的 handler
			// 不写 return 的话，也可以写 runtime.Goexit()
			//runtime.Goexit()
			return
		}
	}

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
