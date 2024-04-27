package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func throw_error(message ...string) {
	fmt.Println(message)
	os.Exit(1)
}

func main() {
	fmt.Printf("Listening to localhost on port 4221...\n\n")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		throw_error("Failed to bind to port 4221")
	}

	args := os.Args[1:]
	for {
		conn, err := l.Accept()
		if err != nil {
			throw_error("Error accepting connection: ", err.Error())
		}
		fmt.Println("\nIncoming connection request from ", conn.RemoteAddr().String())
		go handleConnection(conn, args)
	}
}

func parseHeader(header string) map[string]string {
	header_map := make(map[string]string)
	eachHeader := strings.Split(header, "\r\n")
	startLine := strings.Split(eachHeader[0], " ")
	header_map["method"] = startLine[0]
	header_map["url"] = startLine[1]
	if len(eachHeader) > 2 {
		agent_header := strings.Split(eachHeader[2], ": ")
		if len(agent_header) > 1 {
			header_map["agent"] = agent_header[1]
		}
	}
	return header_map
}

func getHeader(key, value string) string {
	return fmt.Sprintf("%s: %s\r\n", key, value)
}
func appendBody(sendData string, body string) string {
	sendData += getHeader("Content-Type", "text/plain")
	sendData += getHeader("Content-Length", strconv.Itoa(len(body)))
	sendData += "\r\n"
	sendData += body
	return sendData
}

func getStatusLine(status int) string {
	statusLine := fmt.Sprintf("HTTP/1.1 %d ", status)
	switch status {
	case 200:
		statusLine += "OK"
	case 404:
		statusLine += "Not Found"
	}
	statusLine += "\r\n"
	return statusLine
}

func sendData(conn net.Conn, status int, body string) {
	data := getStatusLine(status)
	if body != "" {
		data = appendBody(data, body)
	} else {
		data += "\r\n"
	}
	fmt.Printf("Sending: %s\n", data)
	conn.Write([]byte(data))
}

func sendFile(conn net.Conn, file, directory string) {
	data := getStatusLine(200)
	data += getHeader("Content-Type", "application/octet-stream")

	absolutePath := fmt.Sprintf("%s/%s", directory, file)
	fmt.Println(absolutePath)
	fileData, err := os.ReadFile(absolutePath)
	if err != nil {
		fmt.Println("Could not find the file")
		sendData(conn, 404, "")
		return
	}

	fmt.Println(fileData)
	data += getHeader("Content-Length", strconv.Itoa(len(fileData)))
	data += "\r\n"
	data += string(fileData)
	conn.Write([]byte(data))
}

func getSuffix(url, initial string) string {
	_, after, _ := strings.Cut(url, initial)
	return after
}

func handleURL(conn net.Conn, headerMap map[string]string, args []string) {
	url := headerMap["url"]
	if url == "/" {
		sendData(conn, 200, "")
	} else if strings.HasPrefix(url, "/echo") {
		sendData(conn, 200, getSuffix(url, "/echo/"))
	} else if strings.HasPrefix(url, "/user-agent") {
		sendData(conn, 200, headerMap["agent"])
	} else if strings.HasPrefix(url, "/files") {
		filePath := getSuffix(url, "/files/")
		if len(args) > 1 && args[0] == "--" {
			sendFile(conn, filePath, args[1])
		} else {
			sendData(conn, 404, "")
		}
	} else {
		sendData(conn, 404, "")
	}
}

func handleConnection(conn net.Conn, args []string) {
	connectionBytes := make([]byte, 1024)
	conn.Read(connectionBytes)
	headerMap := parseHeader(string(connectionBytes))
	fmt.Printf("Recieved %s for %s\n\n", headerMap["method"], headerMap["url"])
	handleURL(conn, headerMap, args)
}
