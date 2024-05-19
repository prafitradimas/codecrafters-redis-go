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

var storage = make(map[string]string)

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
		n := len(commands)
		i := 0

		for n > i {
			if strings.Contains(strings.ToLower(commands[i]), PING) {
				conn.Write(encode("+", "PONG"))
			} else if strings.Contains(strings.ToLower(commands[i]), ECHO) {
				conn.Write(encode("$", commands[i+2]))
				i += 2
				continue
			} else if strings.Contains(strings.ToLower(commands[i]), SET) {
				storage[commands[i+2]] = commands[i+4]
				conn.Write(encode("+", "OK"))
				i += 4
				continue
			} else if strings.Contains(strings.ToLower(commands[i]), GET) {
				key := commands[i+2]
				val, found := storage[key]
				if !found {
					val = ""
				}
				conn.Write(encode("$", val))
				i += 2
				continue
			}

			i += 1
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
