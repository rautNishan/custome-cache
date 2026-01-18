// Please take reference https://man7.org/linux/man-pages/man7/epoll.7.html
package core

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rautNishan/custome-cache/common"
	"github.com/rautNishan/custome-cache/config"
	"github.com/rautNishan/custome-cache/protocol"
)

func CreateAndHandelConnection(server *config.Server) {
	//Why this works (https://stackoverflow.com/questions/20424623/proper-way-to-choose-func-at-runtime-for-different-operating-systems)
	mp, err := MultiPlexerInit(server.MaxEvents)
	if err != nil {
		common.PanicOnErr("failed to initialize multiplexer", err)
	}

	//https://stackoverflow.com/questions/5328155/preventing-fin-wait2-when-closing-socket
	//(https://serverfault.com/questions/738300/why-are-connections-in-fin-wait2-state-not-closed-by-the-linux-kernel)
	//Exact error on my pc for command : sudo netstat -tanp
	// (tcp        0      0 127.0.0.1:3000          127.0.0.1:52568         FIN_WAIT2   -)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	defer func() {
		if err := shutDown(mp, server); err != nil {
			fmt.Println("Error during shutdown:", err)
		}
	}()

	err = server.BindAndListen()

	if err != nil {
		common.PanicOnErr("Error", err)
	}
	//Add server fd to intrest list
	err = mp.Add(server.ServerFd)
	if err != nil {
		common.PanicOnErr("failed to add server fd in multiplexer", err)
	}

	for {
		select {
		case sig := <-sigs:
			fmt.Printf("Signal :%s \n", sig)
			return
		default:
			readyEvents, err := mp.Wait()
			if err != nil {
				fmt.Printf("Error in epoll wait: %v\n", err)
				continue
			}
			for _, ev := range readyEvents {
				if ev.Fd == server.ServerFd {
					err := acceptConnAndAddInIntrestedList(server, mp)
					if err != nil {
						fmt.Println(err)
						continue
					}
					fmt.Println("Connected")
				} else {
					tokens, err := readAndGetTokens(ev.Fd)
					if err != nil {
						removeErr := mp.Remove(ev.Fd)
						syscall.Close(ev.Fd)
						if removeErr != nil {
							fmt.Printf("Error removing fd from epoll: %v\n", err)
						}
						log.Println("Client Disconnected")
						continue
					}
					cmd := GetCommand(tokens)
					cmd.EvaluateCmdAndResponde(ev.Fd)
				}
			}
		}

	}
}

func shutDown(mp EventMultiPlexer, server *config.Server) error {
	multiplexerError := mp.Close()
	if multiplexerError != nil {
		return fmt.Errorf("failed to close server fd: %w", multiplexerError)
	}

	serverCloserError := server.CloseServerFd()

	if serverCloserError != nil {
		return fmt.Errorf("failed to close server fd: %w", serverCloserError)
	}
	fmt.Println("Bye ^_^")
	return nil
}

func tokenDecoder(buff []byte) ([]string, error) {
	args, _, err := protocol.Decode(buff)
	if err != nil {
		return nil, err
	}

	argvRaw, ok := args.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected array got %t", args)
	}
	tokens := make([]string, len(argvRaw))
	for i := range argvRaw {
		token, ok := argvRaw[i].(string)
		if !ok {
			return nil, fmt.Errorf("not a string %t", argvRaw[i])
		}
		tokens[i] = token
	}
	return tokens, nil
}

func acceptConnAndAddInIntrestedList(s *config.Server, mp EventMultiPlexer) error {
	clientFileDescriptor, _, err := s.AcceptConnection()
	if err != nil {
		return fmt.Errorf("Error while accepting client: %v", err)
	}
	err = mp.Add(clientFileDescriptor)
	if err != nil {
		return fmt.Errorf("Error while adding client fd to multiplexer: %v", err)
	}
	return nil
}

func readAndGetTokens(fd int) ([]string, error) {
	buf := make([]byte, 1024) //Do not make length of 0
	reader := NewSyscallReader(fd)
	n, err := reader.Read(buf)
	// In this n==0 indicates a graceful shutdown (EFO)
	if n == 0 || err != nil {
		return nil, fmt.Errorf("Error while reading: %v", err)
	}
	tokens, err := tokenDecoder(buf[:n])
	if err != nil {
		return nil, fmt.Errorf("Error while extracting tokens: %v", err)
	}
	return tokens, nil
}
