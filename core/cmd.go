package core

import (
	"errors"
	"fmt"
	"strings"

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
	writer := NewSysCallWriter(fd)
	switch cmd.Command {
	case "COMMAND":
		cmd.evalOK(writer)
	case "PING":
		cmd.evalPing(writer)
	case "INFO":
		cmd.evalOK(writer)
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

func (cmd *Command) evalOK(w Writer) {
	p := protocol.Encode("OK", true)
	w.Write(p)
}
