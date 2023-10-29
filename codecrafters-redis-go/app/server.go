package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func handleError(err error, msg string) {
	if err != nil {
		fmt.Println(msg + ": " + err.Error())
		os.Exit(1)
	}
}

var globalMap = make(map[string]string, 0)
var globalMapMutex sync.Mutex

func getFromGlobalMap(key string) (string, bool) {
	globalMapMutex.Lock()
	defer globalMapMutex.Unlock()

	value, exists := globalMap[key]
	return value, exists
}

func setGlobalMap(key string, value string, expire string) {
	globalMapMutex.Lock()
	defer globalMapMutex.Unlock()
	fmt.Println("Setting: ", key, value)
	if expire != "" {
		expiry, err := strconv.Atoi(expire)
		fmt.Println(expiry)
		handleError(err, "Error converting expiry to int: ")
		time.AfterFunc(time.Millisecond*time.Duration(expiry), func() {
			defer globalMapMutex.Unlock()
			globalMapMutex.Lock()
			delete(globalMap, key)

		})
	}
	globalMap[key] = value
}

func parseCommand(b []byte) ([]string, error) {
	str := string(b)
	// fmt.Println("Parsing: ", str)
	parts := strings.Split(str, "\r\n")
	return parts, nil
}

func executeCommand(cmd []string) (string, error) {
	fmt.Println("Executing: ", cmd)
	switch strings.ToLower(cmd[2]) {
	case "ping":
		return "+PONG\r\n", nil
	case "echo":
		return "+" + cmd[4] + "\r\n", nil
	case "set":
		if len(cmd) > 8 && cmd[8] == "px" {
			setGlobalMap(cmd[4], cmd[6], cmd[10])
		} else {
			setGlobalMap(cmd[4], cmd[6], "")
		}
		return "+OK\r\n", nil
	case "get":
		val, exists := getFromGlobalMap(cmd[4])
		if exists {
			return "+" + val + "\r\n", nil
		} else {
			return "$-1\r\n", nil
		}
	}
	return "scoopidywoop", nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	connBuffer := make([]byte, 65536)

	for {
		bytes, err := conn.Read(connBuffer)
		// fmt.Println(bytes)
		if err != io.EOF {
			// fmt.Println("Received: ", connBuffer[:bytes])
			cmd, err := parseCommand(connBuffer[:bytes])
			resp, err := executeCommand(cmd)
			handleError(err, "Error executing command: ")
			conn.Write([]byte(resp))
		}
		handleError(err, "Error reading: ")
	}
}

func main() {
	fmt.Println("Starting server...")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	handleError(err, "Failed to bind to port 6379")
	defer l.Close()

	for {
		conn, err := l.Accept()
		handleError(err, "Error accepting connection: ")
		fmt.Printf("New connection: %s\n", conn.RemoteAddr().String())
		go handleConn(conn)
	}
}
