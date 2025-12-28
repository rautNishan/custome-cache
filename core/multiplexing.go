// Please take reference https://man7.org/linux/man-pages/man7/epoll.7.html
package core

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/rautNishan/custome-cache/config"
	"github.com/rautNishan/custome-cache/protocol"
)

func CreateAndHandelConnection(cnfg *config.Config) {
	//Why this works (https://stackoverflow.com/questions/20424623/proper-way-to-choose-func-at-runtime-for-different-operating-systems)
	mp, err := MultiPlexerInit(cnfg.MaxEvents)
	if err != nil {
		PanicOnErr("failed to initialize multiplexer", err)
	}
	//Close multiplexer fd
	defer mp.Close()
	sFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		PanicOnErr("failed to create listing file descriptor", err)
	}
	//Close server fd
	defer syscall.Close(sFd)

	err = syscall.SetNonblock(sFd, true)
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
	err = syscall.Bind(sFd, &syscall.SockaddrInet4{
		Port: cnfg.Port,
		Addr: [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
	})

	if err != nil {
		PanicOnErr("bind failed", err)
	}
	err = syscall.Listen(sFd, cnfg.MaxEvents)
	if err != nil {
		PanicOnErr("listen failed", err)
	}
	log.Println("Server listning on host: ", cnfg.Host, "and port: ", cnfg.Port)

	//Add server fd to intrest list
	err = mp.Add(sFd)
	if err != nil {
		PanicOnErr("failed to add server fd in multiplexer", err)
	}

	for {
		readyEvents, err := mp.Wait()
		if err != nil {
			fmt.Printf("Error in epoll wait: %v\n", err)
			continue
		}
		for _, ev := range readyEvents {
			//TODO wrap this into function
			if ev.Fd == sFd {
				clientFileDescriptor, _, err := syscall.Accept(sFd)
				if err != nil {
					fmt.Println("Error while accepting the connection: ", err)
					continue
				}
				syscall.SetNonblock(clientFileDescriptor, true)
				err = mp.Add(clientFileDescriptor)
				if err != nil {
					fmt.Println("Error while adding client fd to multiplexer", err)
					continue
				}
			} else {
				buf := make([]byte, 1024)
				n, err := syscall.Read(int(ev.Fd), buf)
				// In this n==0 indicates a graceful shutdown (EFO)
				if n == 0 || err != nil {
					removeErr := mp.Remove(ev.Fd)
					syscall.Close(ev.Fd)
					if removeErr != nil {
						fmt.Printf("Error removing fd from epoll: %v\n", err)
					}
					log.Println("Client Disconnected")
					continue
				}
				args, read, err := protocol.Decode(buf)
				if err != nil || read == 0 {
					removeErr := mp.Remove(ev.Fd)
					syscall.Close(ev.Fd)
					if removeErr != nil {
						fmt.Printf("Error removing fd from epoll: %v\n", err)
					}
					log.Println("Client Disconnected")
					continue
				}
				fmt.Println(args)
			}
		}

	}
}

// // TODO (abstract)
// func runForLinux(cnfg *config.Config) {
// 	max_client := 10000
// 	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_client)
// 	serverFileDescriptor, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0) //what clients connect to
// 	if err != nil {
// 		PanicOnErr("failed to create listing file descriptor", err)
// 	}
// 	defer syscall.Close(serverFileDescriptor)

// 	err = syscall.SetNonblock(serverFileDescriptor, true)
// 	if err != nil {
// 		PanicOnErr("failed to set non-blocking on listning file descriptor", err)
// 	}

// 	ips, err := net.LookupIP(cnfg.Host)
// 	if err != nil {
// 		PanicOnErr("DNS lookup failed", err)
// 	}

// 	var ipv4 net.IP
// 	for _, ip := range ips {
// 		if v4 := ip.To4(); v4 != nil {
// 			ipv4 = v4
// 			break
// 		}
// 	}

// 	if ipv4 == nil {
// 		PanicOnErr("no IPv4 found for host", nil)
// 	}

// 	log.Println("IP: ", ipv4)

// 	err = syscall.Bind(serverFileDescriptor, &syscall.SockaddrInet4{
// 		Port: cnfg.Port,
// 		Addr: [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
// 	})
// 	if err != nil {
// 		PanicOnErr("bind failed", err)
// 	}

// 	err = syscall.Listen(serverFileDescriptor, max_client)
// 	if err != nil {
// 		PanicOnErr("listen failed", err)
// 	}
// 	log.Println("Server listning on host: ", cnfg.Host, "and port: ", cnfg.Port)

// 	//creates a new epoll instance and returns a file descriptor referring to that instance.
// 	epollFileDescriptor, err := syscall.EpollCreate1(0) //monitors the server socket AND all client sockets (This is not a socket)

// 	if err != nil {
// 		PanicOnErr("failed to create epoll fd", err)
// 	}

// 	var socketServerEvents syscall.EpollEvent = syscall.EpollEvent{
// 		Events: syscall.EPOLLIN,
// 		Fd:     int32(serverFileDescriptor), //for a listening socket, "ready to read" means "a client is trying to connect"
// 	}

// 	err = syscall.EpollCtl(epollFileDescriptor, syscall.EPOLL_CTL_ADD, serverFileDescriptor, &socketServerEvents) //registering the listening socket with epoll
// 	if err != nil {
// 		PanicOnErr("failed to read incoming event to the server", err)
// 	}

// 	for {
// 		readyEvents, err := syscall.EpollWait(epollFileDescriptor, events[:], -1)
// 		if err != nil {
// 			fmt.Printf("Error in epoll wait: %v\n", err)
// 			continue
// 		}
// 		for i := 0; i < readyEvents; i++ {
// 			//If client is trying to connect to our server
// 			//Incoming connection
// 			if int(events[i].Fd) == serverFileDescriptor { //Epollin
// 				clientFileDescriptor, _, err := syscall.Accept(serverFileDescriptor)
// 				if err != nil {
// 					fmt.Println("Error while accepting the connection: ", err)
// 					continue
// 				}
// 				syscall.SetNonblock(clientFileDescriptor, true)

// 				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
// 					Events: syscall.EPOLLIN,
// 					Fd:     int32(clientFileDescriptor),
// 				}
// 				err = syscall.EpollCtl(epollFileDescriptor, syscall.EPOLL_CTL_ADD, clientFileDescriptor, &socketClientEvent)
// 				if err != nil {
// 					PanicOnErr("failed to read incoming event to the server", err)
// 				}

// 			} else {
// 				buf := make([]byte, 1024)
// 				n, err := syscall.Read(int(events[i].Fd), buf)
// 				// In this n==0 indicates a graceful shutdown (EFO)
// 				if n == 0 || err != nil {
// 					err = RemoveFromIntrestListAndCloseConnection(epollFileDescriptor, int(events[i].Fd))
// 					if err != nil {
// 						fmt.Printf("Error removing fd from epoll: %v\n", err)
// 					}
// 					continue
// 				}
// 				args, read, err := protocol.Decode(buf)
// 				if err != nil || read == 0 {
// 					err = RemoveFromIntrestListAndCloseConnection(epollFileDescriptor, int(events[i].Fd))
// 					if err != nil {
// 						fmt.Printf("Error removing fd from epoll: %v\n", err)
// 					}
// 					continue
// 				}
// 				fmt.Println(args)
// 			}
// 		}
// 	}

// }
