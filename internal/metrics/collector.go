package metrics

import (
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

type ProcessMetrics struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryBytes uint64  `json:"memory_bytes"`
}

type SystemInfo struct {
	CPUBrand   string  `json:"cpu_name"`
	Cores      int     `json:"cores"`
	TotalRAMGB float64 `json:"total_ram_gb"`
}

func GetProcessMetrics() (ProcessMetrics, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return ProcessMetrics{}, fmt.Errorf("get process: %w", err)
	}

	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		return ProcessMetrics{}, fmt.Errorf("get cpu percent: %w", err)
	}

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		return ProcessMetrics{}, fmt.Errorf("get memory info: %w", err)
	}

	return ProcessMetrics{
		CPUPercent:  cpuPercent,
		MemoryBytes: memInfo.RSS,
	}, nil
}

func GetSystemInfo() (SystemInfo, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return SystemInfo{}, fmt.Errorf("get cpu info: %w", err)
	}
	if len(cpuInfo) == 0 {
		return SystemInfo{}, fmt.Errorf("get cpu info: empty")
	}

	cores, err := cpu.Counts(true)
	if err != nil {
		return SystemInfo{}, fmt.Errorf("get cpu counts: %w", err)
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return SystemInfo{}, fmt.Errorf("get memory info: %w", err)
	}

	return SystemInfo{
		CPUBrand:   cpuInfo[0].ModelName,
		Cores:      cores,
		TotalRAMGB: float64(vm.Total) / (1024 * 1024 * 1024),
	}, nil
}
