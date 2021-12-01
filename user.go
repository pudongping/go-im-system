package main

import (
	"net"
)

type User struct {
	Name string      // 用户名
	Addr string      // 用户客户端地址
	C    chan string // 当前是否有数据回写给客户端
	conn net.Conn    // 用户的连接
}

// 创建一个用户的 API
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	// 启动监听当前 user channel 消息的 goroutine
	go user.ListenMessage()

	return user
}

// 监听当前 user channel 的方法
func (u *User) ListenMessage() {
	for {
		msg := <-u.C

		u.conn.Write([]byte(msg + "\n"))
	}
}
