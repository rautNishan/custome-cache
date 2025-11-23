package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/rautNishan/custome-cache/config"
	"github.com/rautNishan/custome-cache/protocol"
)

func main() {
	startServer()
}

func startServer() {
	config := config.InitializeConfig()
	fmt.Println(config)
	listner, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))

	if err != nil {
		log.Fatal("Error listning to server")
	}
	var totalConnection int64
	for {

		conn, err := listner.Accept()
		log.Printf("New connection: %s\n", conn.RemoteAddr())
		if err != nil {
			fmt.Printf("Error while reading from connect: %v\n", err)
		}
		go handleConnection(conn, &totalConnection)
	}
}

// Per client connection handle
func handleConnection(conn net.Conn, totalConnection *int64) {
	atomic.AddInt64(totalConnection, 1)
	currentCount := atomic.LoadInt64(totalConnection)
	defer func() {
		conn.Close()
		atomic.AddInt64(totalConnection, -1)
		currentCount := atomic.LoadInt64(totalConnection)
		log.Println("Connection disconnected", conn.RemoteAddr(), "Total connections: ", currentCount)
	}()

	log.Println("New connection establisehd: ", conn.RemoteAddr(), "Total connections: ", currentCount)
	buff := make([]byte, 512)

	accumulationBuffer := make([]byte, 0, 4096)

	for {
		n, err := conn.Read(buff)
		if err != nil {
			break
		}
		accumulationBuffer = append(accumulationBuffer, buff[:n]...)
		fmt.Println("Acc buffer: ", string(accumulationBuffer[:]))
		//Inner for loop in case we miss the ending segments (/r/n in the next segment)
		for {
			val, err := protocol.Decode(accumulationBuffer[:])
			if err != nil {
				log.Printf("Error: %v", err)
				break
			}

			if val == nil {
				log.Println("No ending point, data might be coming")
				break
			}

			if len(accumulationBuffer) == 0 {
				break
			}
		}
	}
}
