package store

import "sync"

type Db struct {
	store map[string]interface{}
	mu    sync.Mutex
}

func NewDb() *Db {
	return &Db{
		store: make(map[string]interface{}),
	}
}

func (d *Db) Get(key string) interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()
	value, ok := d.store[key]
	if !ok {
		return "NOT_FOUND"
	}
	return value
}

func (d *Db) Put(key string, value interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.store[key] = value
}
