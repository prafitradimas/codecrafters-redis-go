package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	PING = "ping"
	ECHO = "echo"
	SET  = "set"
	GET  = "get"
)

var storage = make(map[string]any)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		defer conn.Close()
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			conn.Write([]byte(fmt.Sprintf("-Error %s\r\n", err.Error())))
		}

		commands := strings.Split(string(buf), "\r\n")
		for i, v := range commands {
			if strings.Contains(strings.ToLower(v), PING) {
				conn.Write(encode("+", "PONG"))
			} else if strings.Contains(strings.ToLower(v), ECHO) {
				conn.Write(encode("$", commands[i+2]))
			} else if strings.Contains(strings.ToLower(v), SET) {
				slice := strings.Split(v, " ")
				storage[slice[1]] = slice[2]
				conn.Write(encode("+", "OK"))
			} else if strings.Contains(strings.ToLower(v), GET) {
				_, key, _ := strings.Cut(v, " ")
				conn.Write(encode("$", storage[key].(string)))
			}
		}
	}
}

func encode(firstbyte string, str string) []byte {
	s := ""
	switch firstbyte {
	case "+":
		s = fmt.Sprintf("+%s\r\n", str)
	case "$":
		if 0 == len(str) {
			s = "$-1\r\n"
		} else {
			s = fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
		}
	}
	return []byte(s)
}
