// Please take reference https://man7.org/linux/man-pages/man7/epoll.7.html
package core

import (
	"fmt"
	"log"
	"syscall"

	"github.com/rautNishan/custome-cache/common"
	"github.com/rautNishan/custome-cache/config"
	"github.com/rautNishan/custome-cache/protocol"
)

func CreateAndHandelConnection(cnfg *config.Config) {
	//Why this works (https://stackoverflow.com/questions/20424623/proper-way-to-choose-func-at-runtime-for-different-operating-systems)
	mp, err := MultiPlexerInit(cnfg.MaxEvents)
	if err != nil {
		common.PanicOnErr("failed to initialize multiplexer", err)
	}

	//TODO handel os signals for gracefull shutdown so port use error can be avoided
	//(https://serverfault.com/questions/738300/why-are-connections-in-fin-wait2-state-not-closed-by-the-linux-kernel)
	//Exact error on my pc for command : sudo netstat -tanp
	// (tcp        0      0 127.0.0.1:3000          127.0.0.1:52568         FIN_WAIT2   -)
	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	// defer func() {
	// 	multiplexerError := mp.Close()
	// 	if multiplexerError != nil {
	// 		fmt.Println("Error closing multi plexer: ", err)
	// 	}
	// 	serverCloserError := syscall.Close(sFd)
	// 	if serverCloserError != nil {
	// 		fmt.Println("Error closing server: ", err)
	// 	}
	// 	fmt.Println("Bye ^_^")
	// }()

	err = cnfg.BindAndListen()
	defer closeServerFd(cnfg.ServerFd)

	if err != nil {
		common.PanicOnErr("Error", err)
	}
	//Add server fd to intrest list
	err = mp.Add(cnfg.ServerFd)
	if err != nil {
		common.PanicOnErr("failed to add server fd in multiplexer", err)
	}

	for {
		log.Println("Before wait")
		readyEvents, err := mp.Wait()
		log.Println("After Wait")
		if err != nil {
			fmt.Printf("Error in epoll wait: %v\n", err)
			continue
		}
		for _, ev := range readyEvents {
			//TODO wrap this into function
			if ev.Fd == cnfg.ServerFd {
				clientFileDescriptor, _, err := syscall.Accept(cnfg.ServerFd)
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

				argvRaw, ok := args.([]interface{})
				if !ok || len(argvRaw) == 0 {
					continue
				}

				cmd, ok := argvRaw[0].(string)
				if !ok {
					continue
				}

				if cmd == "PING" || cmd == "ping" {
					syscall.Write(ev.Fd, []byte("+PONG\r\n"))
				}
			}
		}

	}
}

func closeServerFd(sFd int) {
	serverCloseError := syscall.Close(sFd)
	if serverCloseError != nil {
		fmt.Println("Error closing server: ", serverCloseError)
	}
}
