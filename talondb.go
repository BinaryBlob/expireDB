package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	CACHE = map[string]string{}
	bind  = flag.String("bind", "127.0.0.1:11211", "Address:port to bind to")
	db    = flag.String("db", "talon.db", "path to database")
)

type CacheItem struct {
	Key   string
	Value []byte
}

func main() {
	// NUmber of cpu's to use
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	ioChannel := make(chan CacheItem)
	go ioHandler(ioChannel)

	listener, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Printf("net.Listen error")
		panic("Error listening on 11211: " + err.Error())
	}

	CACHE = make(map[string]string)

	log.Printf("\x1b[32m [*] Listening on:\x1b[0m %s", *bind)

	// Load the cache saved on disk
	loadCache()

	for {
		netconn, err := listener.Accept()
		if err != nil {
			log.Printf("Listener.Accept() error")
			panic("Accept error: " + err.Error())
		}

		go handleConn(netconn, ioChannel)
	}

}

func ioHandler(cs chan CacheItem) {
	for {
		item := <-cs
		CACHE[item.Key] = string(item.Value)
	}

}

func loadCache() {
	n, err := ioutil.ReadFile(*db)
	if err != nil {
		return
	}

	p := bytes.NewBuffer(n)

	dec := gob.NewDecoder(p)
	err = dec.Decode(&CACHE)
	if err != nil {
		return
	}
	log.Printf("%v", CACHE)
}

func syncCache() {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(CACHE)
	if err != nil {
		log.Printf("Error detected while encoding: %v", err)
		return
	}

	// Write gob object to file

	// open output file
	fo, err := os.Create(*db)
	if err != nil {
		log.Printf("Saving error")
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := fo.Write(b.Bytes()); err != nil {
		log.Printf("Error writing bytes SAVE")
		panic(err)
	}
}

func handleConn(conn net.Conn, ioHandler chan CacheItem) {
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
				_, err := conn.Write([]uint8("VALUE " + key + " " + string(value) + "\r\n\r\n"))
				if err != nil {
					return
				}
			} else {
				_, err = conn.Write([]uint8("VALUE nil"))
				if err != nil {
					return
				}

			}

			conn.Write([]uint8("\r\n"))
			return

		case "set":
			key := parts[1]

			length := utf8.RuneCountInString(parts[2]) + 120
			val := make([]byte, length)
			val = []byte(parts[2])

			kv := CacheItem{key, val}
			ioHandler <- kv

			_, err := conn.Write([]uint8("STORED\r\n"))
			if err != nil {
				conn.Write([]uint8("ERROR"))
				return
			}
			return

		case "save":
			log.Printf(" [*] Writing CACHE to disk")
			go syncCache()

		case "delete":
			key := parts[1]
			delete(CACHE, key)
			log.Printf(" [*] Deleted [%v] from CACHE", key)
			return

		case "stats":
			stats := strconv.Itoa(len(CACHE))
			_, err = conn.Write([]uint8("KEYS " + stats))
			if err != nil {
				return
			}
		}
	}
}
