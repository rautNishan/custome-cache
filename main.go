package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

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
	for {
		conn, err := listner.Accept()
		log.Printf("New connection: %s\n", conn.RemoteAddr())
		if err != nil {
			fmt.Printf("Error while reading from connect: %v\n", err)
		}
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		log.Printf("Connection closed: %s )\n", conn.RemoteAddr())
	}()

	buff := make([]byte, 512)

	accumulationBuffer := make([]byte, 0, 4096)

	for {
		n, err := conn.Read(buff)
		if err != nil {
			conn.Close()
			log.Println("Connection disconnected", conn.RemoteAddr())
			break
		}
		accumulationBuffer = append(accumulationBuffer, buff[:n]...)
		fmt.Println("Acc buffer: ", string(accumulationBuffer[:]))
		// Decode one
		for {
			val, totalRead, err := protocol.DecodeOne(accumulationBuffer[:])
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
			log.Println(totalRead)
			if totalRead == 0 {
				break
			}

			if val != "" {
				log.Println("Decoded: ", val)
				break
			}

			if len(accumulationBuffer) == 0 {
				break
			}
		}
	}
}
