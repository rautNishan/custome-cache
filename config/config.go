package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"
)

type Server struct {
	Host      string
	Port      int
	MaxEvents int
	ServerFd  int
}

func InitializeConfig() Server {
	server := Server{}

	server.Host = os.Getenv("HOST")
	portStr := os.Getenv("PORT")

	if server.Host == "" {
		server.Host = "localhost"
	}

	if portStr == "" {
		portStr = "3000"
	}
	intPort, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("Error while initializeing port")
	}
	server.Port = intPort

	if server.MaxEvents == 0 {
		server.MaxEvents = 1000
	}

	return server
}

func (server *Server) BindAndListen() error {
	sFd, sErr := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	//If creating is failed no point on closing
	if sErr != nil {
		return fmt.Errorf("failed to create listing socket: %w", sErr)
	}

	var err error

	defer func() {
		if err != nil {
			serverCloseError := syscall.Close(sFd)
			if serverCloseError != nil {
				fmt.Println("Error closing server: ", serverCloseError)
			}
		}
	}()

	err = syscall.SetNonblock(sFd, true)

	if err != nil {
		return fmt.Errorf("failed to set non-blocking on listning file descriptor", err)
	}

	ips, err := net.LookupIP(server.Host)
	if err != nil {
		return fmt.Errorf("DNS lookup failed: %w", err)
	}

	var ipv4 net.IP
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			ipv4 = v4
			break
		}
	}

	if ipv4 == nil {
		return fmt.Errorf("no IPv4 found for host: %s", server.Host)
	}

	err = syscall.Bind(sFd, &syscall.SockaddrInet4{
		Port: server.Port,
		Addr: [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
	})

	if err != nil {
		return fmt.Errorf("bind failed: %w", err)
	}
	err = syscall.Listen(sFd, server.MaxEvents)
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}
	log.Println("Server listning on host: ", server.Host, "and port: ", server.Port)

	server.ServerFd = sFd
	return nil

}

func (s *Server) CloseServerFd() error {
	if err := syscall.Close(s.ServerFd); err != nil {
		return fmt.Errorf("failed to close server fd: %w", err)
	}
	return nil
}
