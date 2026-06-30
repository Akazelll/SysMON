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
			metrics := monitor.GetCurrentMetrics()
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
			d.ramLabel.SetText(fmt.Sprintf("%.1f GB\n(%.1f%%)", metrics.RAMUsedGB, metrics.RAMUsage))
			d.diskLabel.SetText(fmt.Sprintf("%.1f %%", metrics.DiskUsage))
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
			procTable.UpdateData(monitor.GetTopProcesses())
			time.Sleep(3 * time.Second)
		}
	}()
}
