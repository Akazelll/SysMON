package models

type SystemMetric struct {
	CPUUsage float64
	RAMUsage float64
	DiskUsage float64
	NetRXSpeed float64 
	NetTXSpeed float64
	RAMUsedGB  float64
	RAMTotalGB float64
}
type ProcessStat struct {
	PID      int32
	Name     string
	CPUUsage float64
	RAMUsage float32
	RAMUsedGB float64
	RAMTotalGB float64
}