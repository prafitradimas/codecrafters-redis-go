package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	PING = "ping"
	ECHO = "echo"
	SET  = "set"
	GET  = "get"
	PX   = "px"
)

var (
	storage = make(map[string]string)
	since   = make(map[string]time.Time)
	expire  = make(map[string]int)
)

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

		inputs := strings.Split(string(buf), "\r\n")
		command := strings.ToLower(inputs[2])
		if command == PING {
			conn.Write([]byte("+PONG\r\n"))
		} else if command == ECHO {
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(inputs[4]), inputs[4])))
		} else if command == SET {
			key := inputs[4]
			value := inputs[6]
			exp := 0

			if len(inputs) > 8 {
				atoi, _ := strconv.Atoi(inputs[10])
				exp = atoi
			}
			handleSet(key, value, exp)
			conn.Write([]byte("+OK\r\n"))
		} else if command == GET {
			result := handleGet(inputs[4])
			if result == "" {
				conn.Write([]byte("$-1\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(result), result)))
			}
		}
	}
}

func handleGet(key string) string {
	data, found1 := storage[key]
	if found1 {
		exp, found2 := expire[key]
		sincetime := since[key]
		if found2 && time.Now().After(sincetime.Add(time.Duration(exp*int(time.Millisecond)))) {
			delete(storage, key)
			delete(since, key)
			delete(expire, key)

			return ""
		}
	}

	return data
}

func handleSet(key string, value string, exp int) string {
	storage[key] = value
	since[key] = time.Now()
	if exp != 0 {
		expire[key] = exp
	}

	return "OK"
}
