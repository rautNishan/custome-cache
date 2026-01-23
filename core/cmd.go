package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rautNishan/custome-cache/protocol"
)

type CommandType string

const (
	CmdCommand CommandType = "COMMAND"
	CmdPing    CommandType = "PING"
	CmdInfo    CommandType = "INFO"
	CmdSet     CommandType = "SET"
	CmdGet     CommandType = "GET"
	CmdDel     CommandType = "DEL"
	CmdTTL     CommandType = "TTL"
	CmdExpire  CommandType = "EXPIRE"
)

type CommandExecutor interface {
	Execute(args []string, w Writer)
}

var commandRegistry = make(map[CommandType]CommandExecutor)

func RegisterCommand(cmdType CommandType, executro CommandExecutor) {
	commandRegistry[cmdType] = executro
}

type Command struct {
	Command  CommandType
	Args     []string
	executor CommandExecutor
}

func GetCommand(tokens []string) Command {
	cmdType := CommandType(strings.ToUpper(tokens[0]))
	return Command{
		Command:  cmdType,
		Args:     tokens[1:],
		executor: commandRegistry[cmdType],
	}
}

func writeError(msg string, w Writer) {
	err := getErrorMessage(msg)
	p := protocol.Encode(err, false)
	w.Write(p)
}

func (cmd *Command) EvaluateCmdAndResponde(fd int) {
	writer := NewSysCallWriter(fd)
	cmd.executor.Execute(cmd.Args, writer)
}

// Eval Ping
type EvalPingCommand struct {
}

func (c *EvalPingCommand) Execute(args []string, w Writer) {
	if len(args) > 1 {
		writeError("incorrect arguments in ping", w)
	} else if len(args) == 1 {
		p := protocol.Encode(args[0], true)
		w.Write(p)
	} else {
		p := protocol.Encode("PONG", true)
		w.Write(p)
	}
}

// Eval Ok
type EvalOkCommand struct {
}

func (c *EvalOkCommand) Execute(args []string, w Writer) {
	p := protocol.Encode("OK", true)
	w.Write(p)
}

// Eval Set

type EvalSetCommand struct {
}

func (c *EvalSetCommand) Execute(args []string, w Writer) {
	if len(args) <= 1 {
		writeError("bad arguments in set", w)
		return
	}
	var ttlMs int64 = -1 //there was never ttl
	key, val := args[0], args[1]
	//right now only supported ttl
	for i := 2; i < len(args); i++ {
		opt := args[i]
		switch opt {
		//Only set the key if it does not already exist.
		case "NX", "nx":
			i++
			if i != len(args) {
				writeError("bad arguments", w)
				return
			}
			//First check if key exists or not
			exists := storage.Get(key)
			if exists != nil && exists.expiresAt == -1 {
				p := protocol.Encode(nil, false)
				w.Write(p)
				return //No point to check other args
			}

		//Only set the key if it already exists.
		case "XX", "xx":
			i++
			if i != len(args) {
				writeError("bad arguments", w)
				return
			}
			//First check if key exists or not
			exists := storage.Get(key)
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
			if i == len(args) {
				writeError("bad arguments", w)
				return
			}
			ttlSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				writeError("should be an integer", w)
			}
			ttlMs = ttlSec * 1000

		default:
			writeError("bad args", w)
			return
		}

	}
	storage.Put(key, NewEntry(val, ttlMs))
	p := protocol.Encode("OK", true)
	w.Write(p)
}

type EvalGetCommand struct {
}

func (c *EvalGetCommand) Execute(args []string, w Writer) {
	if len(args) != 1 {
		writeError("should be exact one arguments", w)
		return
	}
	key := args[0]
	data := storage.Get(key)
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

type EvalDelCommand struct {
}

func (c *EvalDelCommand) Execute(args []string, w Writer) {
	if len(args) == 0 {
		writeError("should be exact one arguments", w)
		return
	}
	nDeleted := 0
	for i := 0; i < len(args); i++ {
		key := args[i]
		item := storage.Get(key)

		if item == nil {
			continue
		}
		empty := storage.Delete(key)
		if empty != nil { //Deleted unsuccessful
			continue
		}
		nDeleted++
	}
	p := protocol.Encode(nDeleted, false)
	w.Write(p)
}

type EvalTTLCommand struct {
}

func (c *EvalTTLCommand) Execute(args []string, w Writer) {
	if len(args) != 1 {
		writeError("should be exact one arguments", w)
		return
	}
	key := args[0]
	//First check if value exist or not
	item := storage.Get(key)
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

type EvalExpireCommand struct{}

func (c *EvalExpireCommand) Execute(args []string, w Writer) {
	if len(args) != 2 {
		writeError("bad arguments for EXPIRE command", w)
		return
	}

	key := args[0]
	sec, err := strconv.ParseInt(args[1], 10, 64)

	if err != nil {
		writeError("invalid int for expiration", w)
		return
	}
	//First get key
	item := storage.Get(key)

	if item == nil {
		p := protocol.Encode(0, false)
		w.Write(p)
		return
	}
	fmt.Println("Setting new expiration")
	//Set new Expiration time
	item.expiresAt = (sec * 1000) + time.Now().UnixMilli()
	p := protocol.Encode(1, false)
	w.Write(p)
}

type OKCommand struct{}

func (c *OKCommand) Execute(args []string, w Writer) {
	p := protocol.Encode("OK", true)
	w.Write(p)
}

func getErrorMessage(msg string) error {
	return errors.New(msg)
}

func init() {
	RegisterCommand(CmdCommand, &OKCommand{})
	RegisterCommand(CmdPing, &EvalPingCommand{})
	RegisterCommand(CmdInfo, &OKCommand{})
	RegisterCommand(CmdSet, &EvalSetCommand{})
	RegisterCommand(CmdGet, &EvalGetCommand{})
	RegisterCommand(CmdDel, &EvalDelCommand{})
	RegisterCommand(CmdTTL, &EvalTTLCommand{})
	RegisterCommand(CmdExpire, &EvalExpireCommand{})
}
