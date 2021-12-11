package main

import (
	"fmt"
)

func main() {
	ip := "127.0.0.1"
	port := 8888
	server := NewServer(ip, port)
	fmt.Printf("server started at: %s:%d\n", ip, port)
	server.Start()
}
