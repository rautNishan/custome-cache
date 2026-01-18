package core

import (
	"fmt"
	"time"
)

var storage map[string]*Entry

type Entry struct {
	value     interface{}
	expiresAt int64 // milliseconds
}

func init() {
	storage = make(map[string]*Entry)
}

func NewEntry(value interface{}, ttlMs int64) *Entry {
	var defaultExp int64 = -1 //never expires

	if ttlMs > 0 {
		defaultExp = time.Now().UnixMilli() + ttlMs
	}

	return &Entry{
		value:     value,
		expiresAt: defaultExp,
	}
}

func Put(k string, val *Entry) {
	fmt.Println("puttng new entry: ", val)
	storage[k] = val
}

func Get(k string) *Entry {
	return storage[k]
}
