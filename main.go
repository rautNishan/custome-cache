package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/rautNishan/custome-cache/config"
)

func main() {
	startServer()
}

func startServer() {
	config := config.InitializeConfig()
	fmt.Println(config)
	fmt.Println("Before listner")
	listner, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	fmt.Println("After listner")

	if err != nil {
		log.Fatal("Error listning to server")
	}
	fmt.Println("Before Accept")
	conn, err := listner.Accept()
	fmt.Println("After Accept")
	log.Printf("New connection: %s\n", conn.RemoteAddr())
	if err != nil {
		fmt.Printf("Error while reading from connect: %v\n", err)
	}

	buff := make([]byte, 1024)
	fmt.Println("Before Read")
	n, err := conn.Read(buff)
	fmt.Println("After Read")
	if err != nil {
		fmt.Printf("Error while reading from connect: %v\n", err)
	}
	fmt.Printf("Data: %s", string(buff))
	conn.Write([]byte("PONG"))

	fmt.Printf("Len %d\n", n)
}
