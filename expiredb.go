package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Value struct {
	Value  []byte
	expire int
}

var (
	CACHE = map[string]Value{}
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", "127.0.0.1:11211")
	if err != nil {
		panic("Error listening on 11211: " + err.Error())
	}

	CACHE = make(map[string]Value)

	log.Printf("\x1b[32m [*] Listening on:\x1b[0m 127.0.0.1:11211")

	for {
		netconn, err := listener.Accept()
		if err != nil {
			panic("Accept error: " + err.Error())
		}

		go handleConn(netconn)
	}

}

func handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {

		// Fetch
		content, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
			return
		}

		content = content[:len(content)-2] // Chop \r\n

		// Handle
		parts := strings.Split(content, " ")
		cmd := parts[0]

		switch cmd {
		case "get":
			key := parts[1]
			value, ok := CACHE[key]
			if ok {
				conn.Write([]uint8("VALUE " + key + " " + string(value.Value) + "\r\n"))
			}
			conn.Write([]uint8("END\r\n"))

		case "set":
			key := parts[1]

			length := utf8.RuneCountInString(parts[2])
			val := make([]byte, length)
			val = []byte(parts[2])

			expire, _ := strconv.Atoi(parts[3])

			CACHE[key] = Value{val, expire}

			log.Printf(" [*] Stored key")
			conn.Write([]uint8("STORED\r\n"))
		}
	}
}
