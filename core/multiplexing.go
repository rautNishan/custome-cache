// Please take reference https://man7.org/linux/man-pages/man7/epoll.7.html
package core

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/rautNishan/custome-cache/config"
)

func CreateAndHandelConnection(cnfg *config.Config) {
	currentlyUserOs := DetectOS()
	log.Println("Client os: ", currentlyUserOs) //For debugging
	switch currentlyUserOs {
	case 1:
		runForLinux(cnfg)
	}
}

func runForLinux(cnfg *config.Config) {
	max_client := 10000
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_client)
	serverFileDescriptor, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0) //what clients connect to
	if err != nil {
		PanicOnErr("failed to create listing file descriptor", err)
	}
	defer syscall.Close(serverFileDescriptor)

	err = syscall.SetNonblock(serverFileDescriptor, true)
	if err != nil {
		PanicOnErr("failed to set non-blocking on listning file descriptor", err)
	}

	ips, err := net.LookupIP(cnfg.Host)
	if err != nil {
		PanicOnErr("DNS lookup failed", err)
	}

	var ipv4 net.IP
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			ipv4 = v4
			break
		}
	}

	if ipv4 == nil {
		PanicOnErr("no IPv4 found for host", nil)
	}

	log.Println("IP: ", ipv4)

	err = syscall.Bind(serverFileDescriptor, &syscall.SockaddrInet4{
		Port: cnfg.Port,
		Addr: [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
	})
	if err != nil {
		PanicOnErr("bind failed", err)
	}

	err = syscall.Listen(serverFileDescriptor, max_client)
	if err != nil {
		PanicOnErr("listen failed", err)
	}
	log.Println("Server listning on host: ", cnfg.Host, "and port: ", cnfg.Port)

	//creates a new epoll instance and returns a file descriptor referring to that instance.
	epollFileDescriptor, err := syscall.EpollCreate1(0) //monitors the server socket AND all client sockets (This is not a socket)

	if err != nil {
		PanicOnErr("failed to create epoll fd", err)
	}

	var socketServerEvents syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFileDescriptor), //for a listening socket, "ready to read" means "a client is trying to connect"
	}

	err = syscall.EpollCtl(epollFileDescriptor, syscall.EPOLL_CTL_ADD, serverFileDescriptor, &socketServerEvents) //registering the listening socket with epoll
	if err != nil {
		PanicOnErr("failed to read incoming event to the server", err)
	}

	for {
		fmt.Println("Before wait")
		readyEvents, err := syscall.EpollWait(epollFileDescriptor, events[:], -1)
		fmt.Println("After wait")
		if err != nil {
			fmt.Printf("Error in epoll wait: %v\n", err)
			continue
		}

		for i := 0; i < readyEvents; i++ {
			fmt.Println("Inside ready events")
			//If client is trying to connect to our server
			//Incoming connection
			if int(events[i].Fd) == serverFileDescriptor { //Epollin
				clientFileDescriptor, socketAddress, err := syscall.Accept(serverFileDescriptor)
				if err != nil {
					fmt.Println("Error while accepting the connection: ", err)
					continue
				}
				fmt.Println("This is socket address: ", socketAddress)
				syscall.SetNonblock(clientFileDescriptor, true)

				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(clientFileDescriptor),
				}
				err = syscall.EpollCtl(epollFileDescriptor, syscall.EPOLL_CTL_ADD, clientFileDescriptor, &socketClientEvent)
				if err != nil {
					PanicOnErr("failed to read incoming event to the server", err)
				}

			} else {
				buf := make([]byte, 1024)
				n, err := syscall.Read(int(events[i].Fd), buf)
				if err != nil {
					fmt.Println("read error:", err)
					syscall.Close(int(events[i].Fd))
					continue
				}

				if n == 0 {
					fmt.Println("client disconnected")
					syscall.Close(int(events[i].Fd))
					continue
				}
				fmt.Println("Received:", string(buf[:n]))
			}
		}
	}

}
