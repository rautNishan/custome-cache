package command

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"github.com/rautNishan/custome-cache/protocol"
)

type Writer interface {
	Write(data []byte) error
}

type SyscallWriter struct {
	fd int
}

type Command struct {
	Command string
	Args    []string
}

func NewSysCallWriter(fd int) *SyscallWriter {
	return &SyscallWriter{fd: fd}
}
func (w *SyscallWriter) Write(data []byte) error {
	_, err := syscall.Write(w.fd, data)
	return err
}

func GetCommand(tokens []string) Command {
	return Command{
		Command: strings.ToUpper(tokens[0]),
		Args:    tokens[1:],
	}
}

func (cmd *Command) EvaluateCmdAndResponde(fd int) {
	fmt.Println(cmd.Command)
	writer := NewSysCallWriter(fd)
	switch cmd.Command {
	case "PING":
		cmd.evalPing(writer)
	case "INFO":
		cmd.evalInfo(writer)
	default:
		cmd.evalError(writer)
	}
}

func (cmd *Command) evalPing(w Writer) {
	if len(cmd.Args) > 1 {
		err := errors.New("ERR unknown command")
		p := protocol.Encode(err, false)
		w.Write(p)
	} else if len(cmd.Args) == 1 {
		fmt.Println("Yes one", cmd.Args[0])
		p := protocol.Encode(cmd.Args[0], true)
		w.Write(p)

	} else {
		p := protocol.Encode("PONG", true)
		w.Write(p)
	}
}
func (cmd *Command) evalError(w Writer) {
	err := errors.New("unsupported RESP encode type")
	p := protocol.Encode(err, false)
	w.Write(p)
}

func (cmd *Command) evalInfo(w Writer) {
	p := protocol.Encode("OK", true)
	w.Write(p)
}
