package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// comm := "*1\r\n$4\r\nping\r\n"
	comm := "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"
	_, err = conn.Write([]byte(comm))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	response := string(buf[:n])
	fmt.Println(response)
}
