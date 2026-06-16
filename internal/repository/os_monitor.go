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
	cpuPercents, _ := cpu.Percent(time.Second, false)
	cpuVal := 0.0
	if len(cpuPercents) > 0 {
		cpuVal = cpuPercents[0]
	}

	vMem, _ := mem.VirtualMemory()
	ramVal := 0.0
	ramUsed := 0.0
	ramTotal := 0.0
	if vMem != nil {
		ramVal = vMem.UsedPercent
		ramUsed = float64(vMem.Used) / (1024 * 1024 * 1024)
		ramTotal = float64(vMem.Total) / (1024 * 1024 * 1024)
	}

	diskStat, _ := disk.Usage("/")
	diskVal := 0.0
	if diskStat != nil {
		diskVal = diskStat.UsedPercent
	}

	netStats, _ := net.IOCounters(false)
	var rxSpeed, txSpeed float64

	if len(netStats) > 0 {
		currentStat := netStats[0]
		now := time.Now()
		elapsed := now.Sub(m.lastTime).Seconds()

		if elapsed > 0 {
			rxSpeed = float64(currentStat.BytesRecv-m.lastNetStat.BytesRecv) / elapsed
			txSpeed = float64(currentStat.BytesSent-m.lastNetStat.BytesSent) / elapsed
		}

		m.lastNetStat = currentStat
		m.lastTime = now
	}

	return models.SystemMetric{
		CPUUsage:   cpuVal,
		RAMUsage:   ramVal,
		DiskUsage:  diskVal,
		RAMUsedGB: ramUsed,
		RAMTotalGB: ramTotal,
		NetRXSpeed: rxSpeed,
		NetTXSpeed: txSpeed,
	}
}
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

		if name != "" {
			results = append(results, models.ProcessStat{
				PID:      p.Pid,
				Name:     name,
				CPUUsage: cpuPercent,
				RAMUsage: memPercent,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CPUUsage > results[j].CPUUsage
	})

	if len(results) > 15 {
		return results[:15]
	}
	return results
}
func (m *OSMonitor) KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill() 
}