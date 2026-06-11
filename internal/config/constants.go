package config

import "runtime"

const (
	GlobalSeed      int64   = 42
	BatchSize       int     = 4096
	ValidationSplit float64 = 0.2
)

var NumWorkers = runtime.NumCPU()
