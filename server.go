package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sengMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sengMsg
}

func (this *Server) Handler(conn net.Conn) {
	// 链接当前业务
	// fmt.Println("链接建立成功")

	user := NewUser(conn)
	// 用户上线，将用户加入到onlinemap中，广播当前用户消息
	this.mapLock.Lock() //多线程同时操作公共区域（全局变量）要加锁
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	this.BroadCast(user, "已上线")

	// 当前handle goroutine阻塞
	select {} //不阻塞这里user就没有了，map中存的是user指针，值也会没有
}

func (this *Server) Start() {
	//sockert listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听message的goroutine
	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listenr accept err:", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}
}
