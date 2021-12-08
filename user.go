package main

import (
	"net"
)

type User struct {
	Name string      // 用户名
	Addr string      // 用户客户端地址
	C    chan string // 当前是否有数据回写给客户端
	conn net.Conn    // 用户的连接
	server *Server
}

// 创建一个用户的 API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
		server: server,
	}

	// 启动监听当前 user channel 消息的 goroutine
	go user.ListenMessage()

	return user
}

// 用户的上线业务
func (u *User) Online()  {
	// 用户上线，将用户加入到 onlineMap 中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播当前用户上线消息
	u.server.BroadCast(u, "用户已上线")
}

// 用户的下线业务
func (u *User) Offline()  {
	// 用户下线，将用户从 onlineMap 中删除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播当前用户下线消息
	u.server.BroadCast(u, "用户已下线")
}

// 用户处理消息的业务
func (u *User) DoMessage(msg string)  {
	// 将得到的消息进行广播
	u.server.BroadCast(u, msg)
}

// 监听当前 user channel 的方法
func (u *User) ListenMessage() {
	for {
		msg := <-u.C

		u.conn.Write([]byte(msg + "\n"))
	}
}
