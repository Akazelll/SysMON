package repository

import (
	"sort"
	"time"

	"sysmon-app/internal/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type OSMonitor struct {
	lastNetStat net.IOCountersStat
	lastTime    time.Time
}

func NewOSMonitor() *OSMonitor {
	// Ambil state network awal sebagai titik mulai
	netStats, _ := net.IOCounters(false)
	initialStat := net.IOCountersStat{}
	if len(netStats) > 0 {
		initialStat = netStats[0]
	}

	return &OSMonitor{
		lastNetStat: initialStat,
		lastTime:    time.Now(),
	}
}

func (m *OSMonitor) GetCurrentMetrics() models.SystemMetric {
	// 1. CPU & RAM
	cpuPercents, _ := cpu.Percent(time.Second, false)
	cpuVal := 0.0
	if len(cpuPercents) > 0 {
		cpuVal = cpuPercents[0]
	}

	vMem, _ := mem.VirtualMemory()
	ramVal := 0.0
	if vMem != nil {
		ramVal = vMem.UsedPercent
	}

	// 2. DISK (Ambil penggunaan partisi root/C:)
	diskStat, _ := disk.Usage("/")
	diskVal := 0.0
	if diskStat != nil {
		diskVal = diskStat.UsedPercent
	}

	// 3. NETWORK (Hitung Throughput RX & TX)
	netStats, _ := net.IOCounters(false)
	var rxSpeed, txSpeed float64

	if len(netStats) > 0 {
		currentStat := netStats[0]
		now := time.Now()
		elapsed := now.Sub(m.lastTime).Seconds()

		if elapsed > 0 {
			// Selisih byte dibagi detik = Bytes/second
			rxSpeed = float64(currentStat.BytesRecv-m.lastNetStat.BytesRecv) / elapsed
			txSpeed = float64(currentStat.BytesSent-m.lastNetStat.BytesSent) / elapsed
		}

		// Simpan state saat ini untuk kalkulasi detik berikutnya
		m.lastNetStat = currentStat
		m.lastTime = now
	}

	return models.SystemMetric{
		CPUUsage:   cpuVal,
		RAMUsage:   ramVal,
		DiskUsage:  diskVal,
		NetRXSpeed: rxSpeed,
		NetTXSpeed: txSpeed,
	}
}
// GetTopProcesses mengambil daftar proses dan mengurutkannya berdasarkan CPU tertinggi
func (m *OSMonitor) GetTopProcesses() []models.ProcessStat {
	procs, err := process.Processes()
	if err != nil {
		return nil
	}

	var results []models.ProcessStat
	for _, p := range procs {
		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()

		// Filter proses yang kosong atau tidak ada namanya
		if name != "" {
			results = append(results, models.ProcessStat{
				PID:      p.Pid,
				Name:     name,
				CPUUsage: cpuPercent,
				RAMUsage: memPercent,
			})
		}
	}

	// Urutkan dari CPU Usage tertinggi ke terendah (Descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CPUUsage > results[j].CPUUsage
	})

	// Ambil top 15 saja agar aplikasi tetap ringan
	if len(results) > 15 {
		return results[:15]
	}
	return results
}