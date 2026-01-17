package command

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"github.com/rautNishan/custome-cache/protocol"
)

type Command struct {
	Command string
	Args    []string
}

func GetCommand(tokens []string) Command {
	return Command{
		Command: strings.ToUpper(tokens[0]),
		Args:    tokens[1:],
	}
}

func (cmd *Command) EvaluateCmdAndResponde(fd int) {
	fmt.Println(cmd.Command)
	switch cmd.Command {
	case "PING":
		cmd.evalPing(fd)
	case "INFO":
		cmd.evalInfo(fd)
	default:
		cmd.evalError(fd)
	}
}

func (cmd *Command) evalPing(fd int) {
	if len(cmd.Args) > 1 {
		err := errors.New("ERR unknown command")
		p := protocol.Encode(err, false)
		syscall.Write(fd, p)
	} else if len(cmd.Args) == 1 {
		fmt.Println("Yes one", cmd.Args[0])
		p := protocol.Encode(cmd.Args[0], true)
		syscall.Write(fd, p)
	} else {
		p := protocol.Encode("PONG", true)
		syscall.Write(fd, p)
	}
}
func (cmd *Command) evalError(fd int) {
	err := errors.New("unsupported RESP encode type")
	p := protocol.Encode(err, false)
	syscall.Write(fd, p)
}

func (cmd *Command) evalInfo(fd int) {
	p := protocol.Encode("OK", true)
	syscall.Write(fd, p)
}
