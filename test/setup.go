package test

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	RootDir    = filepath.Dir(b)
	ActionLog  = filepath.Join(RootDir, "fixtures", "put.txt")
)
