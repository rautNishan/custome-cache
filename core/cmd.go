package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	case "SET":
		cmd.evalSet(writer)
	case "GET":
		cmd.evalGet(writer)
	case "TTL":
	default:
		cmd.evalError("unsupported RESP encode type", writer)
	}
}

func (cmd *Command) evalPing(w Writer) {
	if len(cmd.Args) > 1 {
		cmd.evalError("incorrect arguments in ping", w)
	} else if len(cmd.Args) == 1 {
		p := protocol.Encode(cmd.Args[0], true)
		w.Write(p)

	} else {
		p := protocol.Encode("PONG", true)
		w.Write(p)
	}
}

func (cmd *Command) evalOK(w Writer) {
	p := protocol.Encode("OK", true)
	w.Write(p)
}

func (cmd *Command) evalSet(w Writer) {
	if len(cmd.Args) <= 1 {
		cmd.evalError("bad arguments in set", w)
		return
	}
	var ttlMs int64 = -1 //there was never ttl
	key, val := cmd.Args[0], cmd.Args[1]

	//right now only supported ttl
	for i := 2; i < len(cmd.Args); i++ {
		ex := cmd.Args[i]
		switch ex {
		case "EX", "ex":
			i++
			//set k v ttl (but no value in ttl)
			if i == len(cmd.Args) {
				cmd.evalError("bad arguments", w)
				return
			}
			ttlSec, err := strconv.ParseInt(cmd.Args[i], 10, 64)
			if err != nil {
				cmd.evalError("should be an integer", w)
			}
			ttlMs = ttlSec * 1000
		default:
			cmd.evalError("bad args", w)
			return
		}

	}
	Put(key, NewEntry(val, ttlMs))
	cmd.evalOK(w)
}

func (cmd *Command) evalGet(w Writer) {
	fmt.Println("Inside get")
	if len(cmd.Args) != 1 {
		cmd.evalError("should be exact one arguments", w)
		return
	}
	data := Get(cmd.Args[0])
	fmt.Println("This is data: ", data)
	//If the key did not exists
	if data == nil {
		p := protocol.Encode(nil, false)
		w.Write(p)
		return
	}

	//check if user has put expiration or not
	fmt.Println("This is data expired at: ", data.expiresAt)
	if data.expiresAt != -1 && data.expiresAt <= time.Now().UnixMilli() {
		data.expiresAt = -2
		p := protocol.Encode(nil, false)
		w.Write(p)
		return
	}

	p := protocol.Encode(data.value, false)
	w.Write(p)
}

func (cmd *Command) evalError(msg string, w Writer) {
	err := getErrorMessage(msg)
	p := protocol.Encode(err, false)
	w.Write(p)
}

func getErrorMessage(msg string) error {
	return errors.New(msg)
}
