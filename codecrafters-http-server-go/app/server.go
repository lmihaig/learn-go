package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Request struct {
	Method    string
	Path      string
	Host      string
	UserAgent string
	Body      string
}

func printStringList(stringsList []string) {
	formattedString := ""

	for i, str := range stringsList {
		formattedString += "\"" + str + "\""

		if i < len(stringsList)-1 {
			formattedString += ", "
		}
	}

	// Print the formatted string
	fmt.Println(formattedString)
}

func stringToRequest(request string) Request {
	rest, body := strings.Split(request, "\r\n\r\n")[0], strings.Split(request, "\r\n\r\n")[1]
	requestParts := strings.Split(rest, "\r\n")
	req := Request{}
	req.Body = body
	req.Method = strings.Split(requestParts[0], " ")[0]
	req.Path = strings.Split(requestParts[0], " ")[1]
	for _, part := range requestParts[1:] {
		if strings.HasPrefix(part, "Host: ") {
			req.Host = strings.TrimPrefix(part, "Host: ")
		} else if strings.HasPrefix(part, "User-Agent: ") {
			req.UserAgent = strings.TrimPrefix(part, "User-Agent: ")
		}
	}
	// printStringList(requestParts)
	return req
}

func handleError(err error, msg string) {
	if err != nil {
		fmt.Println(msg + ": " + err.Error())
	}
}

func handleGet(request Request) string {

	if request.Path == "/" {
		return "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 0\r\n\r\n"
	}
	if strings.HasPrefix(request.Path, "/echo/") {
		content := strings.Split(request.Path, "/echo/")[1]
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
	}
	if strings.HasPrefix(request.Path, "/user-agent") {
		content := request.UserAgent
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
	}
	if strings.HasPrefix(request.Path, "/files/") {
		dir := os.Args[2]
		fmt.Println(dir)
		filePath := strings.Split(request.Path, "/files/")[1]
		content, err := os.ReadFile(dir + "/" + filePath)
		if err != nil {
			return "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)

	}

	return "HTTP/1.1 404 Not Found\r\n\r\n"
}

func handlePost(request Request) string {
	if strings.HasPrefix(request.Path, "/files/") {
		dir := os.Args[2]
		fmt.Println(dir)
		filePath := strings.Split(request.Path, "/files/")[1]
		err := os.WriteFile(dir+"/"+filePath, []byte(request.Body), 0644)
		if err != nil {
			return "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		return fmt.Sprintf("HTTP/1.1 201 Created\r\nContent-Type: application/json\r\nLocation: /users/124 %s\r\n\r\n%d", dir+"/"+filePath, 0)
	}
	return "HTTP/1.1 404 Not Found\r\n\r\n"
}

func handleRequest(request Request) string {
	switch request.Method {
	case "GET":
		return handleGet(request)

	case "POST":
		return handlePost(request)

	case "PUT":
		return "HTTP/1.1 200 OK\r\n\r\n"

	case "DELETE":
		return "HTTP/1.1 200 OK\r\n\r\n"
	}
	return "HTTP/1.1 400 Bad Request\r\n\r\n"
}

func parseRequest(buffer []byte) (Request, error) {
	request := string(buffer)
	if len(request) == 0 {
		return Request{}, fmt.Errorf("Empty request")
	}
	return stringToRequest(request), nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 4096)
	req, err := conn.Read(buffer)

	// fmt.Println("Received request: ", string(buffer[:req]))

	if err != io.EOF {
		handleError(err, "Failed to read from stream: ")
	}
	request, err := parseRequest(buffer[:req])
	if err != nil {
		fmt.Println("Failed to parse request")
		conn.Close()
		// _, err = conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
	} else {
		fmt.Println("Received request: ", request)
		resp := handleRequest(request)
		_, err = conn.Write([]byte(resp))
		handleError(err, "Failed to write to stream: ")
	}
	conn.Close()
}

func main() {
	fmt.Println("Server is starting...")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	handleError(err, "Failed to bind to port 4221")
	defer l.Close()

	for {
		conn, err := l.Accept()
		handleError(err, "Failed to accept connection: ")
		fmt.Printf("New connection: %s\n", conn.RemoteAddr().String())
		go handleConn(conn)
	}
}
