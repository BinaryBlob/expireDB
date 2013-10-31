package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"unicode/utf8"
)

var (
	CACHE = map[string]string{}
	bind  = flag.String("bind", "127.0.0.1:11211", "Address:port to bind to")
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *bind)
	if err != nil {
		panic("Error listening on 11211: " + err.Error())
	}

	CACHE = make(map[string]string)

	log.Printf("\x1b[32m [*] Listening on:\x1b[0m %s", *bind)

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
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
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
			//log.Printf(" [*] GET key")
			key := parts[1]
			value, ok := CACHE[key]
			if ok {
				_, err := conn.Write([]uint8("VALUE " + key + " " + string(value) + "\r\n\r\n"))
				if err != nil {
					return
				}
			} else {
				_, err = conn.Write([]uint8("VALUE none"))
				if err != nil {
					return
				}

			}

			conn.Write([]uint8("\r\n"))

			conn.Close()

		case "set":
			key := parts[1]

			length := utf8.RuneCountInString(parts[2]) + 120
			val := make([]byte, length)
			val = []byte(parts[2])

			CACHE[key] = string(val)

			//log.Printf(" [*] Stored key")
			conn.Write([]uint8("STORED\r\n"))

			//conn.Close()
		}
	}
}
