package models

// SystemMetric adalah representasi data dari state OS kita saat ini.
// Huruf kapital di awal (seperti CPUUsage) berarti properti ini "Public" (bisa diakses file lain).
type SystemMetric struct {
	CPUUsage float64
	RAMUsage float64
	DiskUsage float64
	NetRXSpeed float64 
	NetTXSpeed float64
	RAMUsedGB  float64 // Tambahan
	RAMTotalGB float64
}
// Tambahkan di bagian bawah file internal/models/metric.go
type ProcessStat struct {
	PID      int32
	Name     string
	CPUUsage float64
	RAMUsage float32
	RAMUsedGB float64
	RAMTotalGB float64
}