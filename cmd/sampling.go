package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/models"
	"sysmon-app/internal/repository"
	"sysmon-app/internal/services"
	"sysmon-app/internal/ui"
)

func startSampling(
	monitor *repository.OSMonitor,
	alertEngine *services.AlertEngine,
	cpuBuffer, ramBuffer *models.RingBuffer,
	history *models.MetricHistory,
	d *dashboardUI,
	historyList *widget.Entry,
	procTable *ui.ProcessTable,
) {
	go func() {
		for {
			metrics, err := monitor.GetCurrentMetrics()
			if err != nil {
				continue
			}
			alertEngine.Evaluate(metrics)

			cpuBuffer.Add(metrics.CPUUsage)
			ramBuffer.Add(metrics.RAMUsage)
			history.Add(models.MetricSample{
				Time:       time.Now(),
				CPUUsage:   metrics.CPUUsage,
				RAMUsage:   metrics.RAMUsage,
				DiskUsage:  metrics.DiskUsage,
				NetRXSpeed: metrics.NetRXSpeed,
				NetTXSpeed: metrics.NetTXSpeed,
			})

			d.cpuLabel.SetText(fmt.Sprintf("%.1f %%", metrics.CPUUsage))
			d.ramLabel.SetText(fmt.Sprintf("%.1f GB / %.1f GB\n(%.1f%%)", metrics.RAMUsedGB, metrics.RAMTotalGB, metrics.RAMUsage))

			
			var totalUsedGB, totalCapacityGB float64
			for _, disk := range metrics.Disks {
				totalUsedGB += disk.UsedGB
				totalCapacityGB += disk.TotalGB
			}

			diskText := fmt.Sprintf("%.1f %%", metrics.DiskUsage)
			if totalCapacityGB > 0 {
				aggregatePercent := (totalUsedGB / totalCapacityGB) * 100.0
				diskText = fmt.Sprintf("%.1f GB / %.1f GB\n(%.1f%%)", totalUsedGB, totalCapacityGB, aggregatePercent)
			}
			d.diskLabel.SetText(diskText)

			d.netLabel.SetText(fmt.Sprintf("↓ %.1f\n↑ %.1f", metrics.NetRXSpeed/1024, metrics.NetTXSpeed/1024))

			d.coreView.Update(metrics.CPUPerCore)
			d.diskView.Update(metrics.Disks)
			d.sysChart.Update(cpuBuffer.Data, ramBuffer.Data, 500, 170)
			historyList.SetText(strings.Join(alertEngine.GetHistory(), "\n"))

			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			procs, err := monitor.GetTopProcesses()
			if err == nil {
				procTable.UpdateData(procs)
			}
			time.Sleep(3 * time.Second)
		}
	}()
}
