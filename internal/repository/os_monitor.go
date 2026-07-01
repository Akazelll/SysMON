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

func (m *OSMonitor) GetCurrentMetrics() (models.SystemMetric, error) {
	perCore, err := cpu.Percent(time.Second, true)
	if err != nil {
		return models.SystemMetric{}, err
	}
	cpuVal := 0.0
	if len(perCore) > 0 {
		sum := 0.0
		for _, c := range perCore {
			sum += c
		}
		cpuVal = sum / float64(len(perCore))
	}

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return models.SystemMetric{}, err
	}
	ramVal := vMem.UsedPercent
	ramUsed := float64(vMem.Used) / (1024 * 1024 * 1024)
	ramTotal := float64(vMem.Total) / (1024 * 1024 * 1024)

	disks := m.getDiskPartitions()
	diskVal := 0.0
	if len(disks) > 0 {
		diskVal = disks[0].UsedPercent
	}

	netStats, err := net.IOCounters(false)
	if err != nil {
		return models.SystemMetric{}, err
	}
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
		CPUPerCore: perCore,
		RAMUsage:   ramVal,
		DiskUsage:  diskVal,
		Disks:      disks,
		RAMUsedGB:  ramUsed,
		RAMTotalGB: ramTotal,
		NetRXSpeed: rxSpeed,
		NetTXSpeed: txSpeed,
	}, nil
}

func (m *OSMonitor) getDiskPartitions() []models.DiskPartition {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	const gb = 1024 * 1024 * 1024
	var results []models.DiskPartition
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage == nil || usage.Total == 0 {
			continue
		}
		results = append(results, models.DiskPartition{
			Mountpoint:  p.Mountpoint,
			UsedPercent: usage.UsedPercent,
			UsedGB:      float64(usage.Used) / gb,
			TotalGB:     float64(usage.Total) / gb,
		})
	}
	return results
}
func (m *OSMonitor) GetTopProcesses() ([]models.ProcessStat, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
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
		return results[:15], nil
	}
	return results, nil
}