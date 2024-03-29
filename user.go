package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string      // 用户名
	Addr   string      // 用户客户端地址
	C      chan string // 当前是否有数据回写给客户端
	conn   net.Conn    // 用户的连接
	server *Server
}

// 创建一个用户的 API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听当前 user channel 消息的 goroutine
	go user.ListenMessage()

	return user
}

// 用户的上线业务
func (u *User) Online() {
	// 用户上线，将用户加入到 onlineMap 中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播当前用户上线消息
	u.server.BroadCast(u, "用户已上线")
}

// 用户的下线业务
func (u *User) Offline() {
	// 用户下线，将用户从 onlineMap 中删除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播当前用户下线消息
	u.server.BroadCast(u, "用户已下线")
}

// 给当前 user 对应的客户端发送消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (u *User) DoMessage(msg string) {

	// 当客户端发送 `who` 指令，则查询当前在线的用户有哪些
	if msg == "who" {
		// 消息格式为 who
		// 查询当前在线用户都有哪些
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := fmt.Sprintf("[%s] %s:在线……\n", user.Addr, user.Name)
			// 谁发送的 `who` 指令，将消息发送给谁
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式为： rename|alex
		// 用的新昵称
		newName := strings.Split(msg, "|")[1]

		// 判断 name 是否存在
		if _, ok := u.server.OnlineMap[newName]; ok {
			u.SendMsg("当前用户名已经被使用\n")
		} else {
			u.server.mapLock.Lock()
			// 将老的用户昵称删除掉
			delete(u.server.OnlineMap, u.Name)
			// 添加新的用户昵称
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg(fmt.Sprintf("您已经更新您的用户名，新的用户名为：[%s]\n", newName))
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 发送消息给指定人，消息格式为： to|{用户名}|{消息内容}，比如：to|alex|hello

		// 1、获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("消息格式不正确，请使用 \"to|张三|你好啊\"格式。 \n")
			return
		}

		// 2、根据用户名，得到对方 User 对象
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("该用户名不存在\n")
			return
		}

		// 3、获取消息内容，通过对方的 User 对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("无消息内容，请重发\n")
			return
		}

		// 将消息内容发送给指定用户
		remoteUser.SendMsg(fmt.Sprintf("%s 对您说： %s\n", u.Name, content))

	} else {
		// 将得到的消息进行广播
		u.server.BroadCast(u, msg)
	}

}

// 监听当前 user channel 的方法
func (u *User) ListenMessage() {
	for {
		msg := <-u.C

		u.conn.Write([]byte(msg + "\n"))
	}
}
