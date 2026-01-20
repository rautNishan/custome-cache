package core

import (
	"time"
)

type Cache interface {
	Put(k string, value *Entry)
	Get(k string) *Entry
	Delete(k string) *Entry
}

var storage Cache

type Entry struct {
	value     interface{}
	expiresAt int64 // milliseconds
}

type InMemoryStorage struct {
	storage map[string]*Entry
}

func NewInMemoryCache() *InMemoryStorage {
	return &InMemoryStorage{
		storage: make(map[string]*Entry),
	}
}

func init() {
	storage = NewInMemoryCache()
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

func (c *InMemoryStorage) Put(k string, val *Entry) {
	c.storage[k] = val
}

func (c *InMemoryStorage) Get(k string) *Entry {
	return c.storage[k]
}

func (c *InMemoryStorage) Delete(k string) *Entry {
	delete(c.storage, k)
	return c.storage[k]
}
