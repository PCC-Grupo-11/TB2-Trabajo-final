package config

import "runtime"

const (
	GlobalSeed      = 42
	BatchSize       = 4096
	ValidationSplit = 0.2
)

var NumWorkers = runtime.NumCPU()
