package core

import (
	"errors"
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
	case "DEL":
		cmd.evalDelete(writer)
	case "TTL":
		cmd.evalTTL(writer)
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
		opt := cmd.Args[i]
		switch opt {
		//Only set the key if it does not already exist.
		case "NX", "nx":
			i++
			if i != len(cmd.Args) {
				cmd.evalError("bad arguments", w)
				return
			}
			//First check if key exists or not
			exists := Get(key)
			if exists != nil && exists.expiresAt == -1 {
				p := protocol.Encode(nil, false)
				w.Write(p)
				return //No point to check other args
			}

		//Only set the key if it already exists.
		case "XX", "xx":
			i++
			if i != len(cmd.Args) {
				cmd.evalError("bad arguments", w)
				return
			}
			//First check if key exists or not
			exists := Get(key)
			if exists == nil {
				p := protocol.Encode(nil, false)
				w.Write(p)
				return //No point to check other args
			}

			//If it has already expired we will mark it as key does not exists
			if exists.expiresAt == -2 {
				p := protocol.Encode(nil, false)
				w.Write(p)
				return
			}
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
	if len(cmd.Args) != 1 {
		cmd.evalError("should be exact one arguments", w)
		return
	}
	key := cmd.Args[0]
	data := Get(key)
	//If the key did not exists
	if data == nil {
		p := protocol.Encode(nil, false)
		w.Write(p)
		return
	}

	//check if user has put expiration or not
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

func (cmd *Command) evalDelete(w Writer) {
	if len(cmd.Args) != 1 {
		cmd.evalError("should be exact one arguments", w)
		return
	}
	key := cmd.Args[0]
	//First check if value exist or not
	item := Get(key)

	if item == nil { //No item to delete
		p := protocol.Encode(nil, false)
		w.Write(p)
		return
	}

	empty := Delete(key)
	if empty != nil {
		cmd.evalError("Something went wrong deleting the key", w)
		return
	}
	cmd.evalOK(w)
}

func (cmd *Command) evalTTL(w Writer) {
	if len(cmd.Args) != 1 {
		cmd.evalError("should be exact one arguments", w)
		return
	}
	key := cmd.Args[0]
	//First check if value exist or not
	item := Get(key)
	if item == nil { //No item to gets its ttl
		p := protocol.Encode(nil, false)
		w.Write(p)
		return
	}

	var sec int64 = -1

	if item.expiresAt == -1 {
		p := protocol.Encode(sec, false)
		w.Write(p)
		return
	}

	if item.expiresAt != -1 && item.expiresAt <= time.Now().UnixMilli() {
		p := protocol.Encode(-2, false)
		w.Write(p)
		return
	}

	sec = (item.expiresAt - time.Now().UnixMilli()) / 1000

	p := protocol.Encode(sec, false)
	w.Write(p)
}
