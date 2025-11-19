package main

import (
	"fmt"
	"net"
)

func main() {
	port := ":7777"

	fmt.Println("Listening on", port)

	ln, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		fmt.Println("Got connection from", conn.RemoteAddr())
		conn.Write([]byte("hello from go server\n"))
		conn.Close()
	}
}
