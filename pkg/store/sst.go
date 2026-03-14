package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type SSRecord struct {
	Key   string      `json:"key" yaml:"key" toml:"key"`
	Value interface{} `json:"value" yaml:"value" toml:"value"`
}

func (d *Db) flush(store map[string]interface{}, manifest *Manifest) error {
	sortedKeys := make([]string, 0, len(store))
	for key := range store {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	sstFileName := fmt.Sprintf(filenameTemplate, manifest.NextSSTableId)
	newFileName := fmt.Sprintf("%s/%s", d.StoragePath, sstFileName)
	file, err := os.OpenFile(newFileName, os.O_CREATE|os.O_RDWR, 0644)
	defer file.Close()
	if err != nil {
		return err
	}
	var items []SSRecord
	for _, key := range sortedKeys {
		items = append(items, SSRecord{Key: key, Value: store[key]})
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	manifest.NextSSTableId++
	return manifest.Append(sstFileName)
}
