package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

func main() {
	address := "127.0.0.1:4221"

	req := "GET / HTTP/1.1\r\n" + "\r\n\r\n"

	conn, _ := net.Dial("tcp", address)
	_, _ = conn.Write([]byte(req))

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		fmt.Println("Failed to read response :", err.Error())
	}
	fmt.Println(resp)
}
