package main

import (
	"fmt"
	"strings"
	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func throw_error(message ...string) {
	fmt.Println(message)
	os.Exit(1)
}

func main() {
	fmt.Println("Listening to localhost on port 4221...")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		throw_error("Failed to bind to port 4221")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			throw_error("Error accepting connection: ", err.Error())
		}
		fmt.Println("Incoming connection request from ", conn.RemoteAddr().String())
		go handleConnection(conn)
	}
}

func parseHeader(header string) map[string]string {
	header_map := make(map[string]string)
	eachHeader := strings.Split(header, "\r\n")
	startLine := strings.Split(eachHeader[0], " ")
	header_map["method"] = startLine[0]
	header_map["url"] = startLine[1]
	return header_map
}

func sendData(conn net.Conn, status int) {
	sendData := fmt.Sprintf("HTTP/1.1 %d ", status)
	switch status {
	case 200:
		sendData += "OK"
	case 404:
		sendData += "Not Found"
	default:
		os.Exit(1)
	}

	fmt.Printf("Sending %s to %s", sendData, conn.RemoteAddr().String())
	conn.Write([]byte(sendData))
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	connectionBytes := make([]byte, 1024)
	conn.Read(connectionBytes)
	headerMap := parseHeader(string(connectionBytes))
	fmt.Printf("Recieved %s for %s\n", headerMap["method"], headerMap["url"])

	if headerMap["url"] == "/" {
		sendData(conn, 200)
	} else {
		sendData(conn, 404)
	}

}
