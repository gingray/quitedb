package store

import (
	"context"
	"sync"

	"github.com/gingray/quitedb/pkg/config"
)

const flushCount = 2000

type Db struct {
	StoragePath           string
	Manifest              *Manifest
	store                 map[string]interface{}
	currentOperationCount int
	mu                    sync.Mutex
}

func (d *Db) Name() string {
	return "db"
}

func (d *Db) Ready(ctx context.Context) error {
	return d.Manifest.InitManifest()
}

func (d *Db) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *Db) Shutdown(ctx context.Context) error {
	return d.Manifest.Close()
}

func NewDb(cfg *config.StorageConfig) *Db {
	return &Db{
		StoragePath:           cfg.StoragePath,
		Manifest:              NewManifest(cfg.StoragePath),
		store:                 make(map[string]interface{}),
		currentOperationCount: 0,
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
	d.currentOperationCount++
	if d.currentOperationCount >= flushCount {
		err := d.flush(d.store, d.Manifest)
		if err != nil {
			panic(err)
		}
		d.currentOperationCount = 0
	}
	d.store[key] = value
}
