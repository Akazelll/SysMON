package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/models"
	"sysmon-app/internal/repository"
	"sysmon-app/internal/services"
	"sysmon-app/internal/ui"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("System Resource Monitor Pro")

	// --- SETUP DATA & BACKEND ---
	monitor := repository.NewOSMonitor()
	alertEngine := services.NewAlertEngine(myApp)
	cpuBuffer := models.NewRingBuffer(60)
	ramBuffer := models.NewRingBuffer(60)

	// --- TAB 1: DASHBOARD ---
	cpuLabel := widget.NewLabel("CPU: - %")
	ramLabel := widget.NewLabel("RAM: - / - GB (- %)")
	diskLabel := widget.NewLabel("Disk: - %")
	netLabel := widget.NewLabel("Net: RX - | TX - KB/s")

	sysChart := ui.NewSystemChart()
	chartContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(500, 150)), sysChart.Container)
	
	exporter := services.NewExporter()
	exportBtn := widget.NewButton("Export CPU CSV", func() {
		dialog.ShowFileSave(func(w fyne.URIWriteCloser, e error) {
			if w == nil { return }
			exporter.ExportCPUHistory(w, cpuBuffer.Data)
			dialog.ShowInformation("Sukses", "Data riwayat CPU berhasil diexport!", myWindow)
		}, myWindow)
	})

	dashboardTab := container.NewVBox(
		exportBtn,
		container.NewGridWithColumns(2, cpuLabel, ramLabel),
		container.NewGridWithColumns(2, diskLabel, netLabel),
		widget.NewLabelWithStyle("Live Performance (60s)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		chartContainer,
		widget.NewLabel("🟥 CPU Usage (%) | 🟦 RAM Usage (%)"),
	)

	// --- TAB 2: PROCESSES ---
	procTable := ui.NewProcessTable()
	sortSelect := widget.NewSelect([]string{"CPU", "RAM"}, func(v string) {
		procTable.SortBy = v
		procTable.Table.Refresh()
	})
	sortSelect.SetSelected("CPU")
	
	processTab := container.NewBorder(
		container.NewHBox(widget.NewLabel("Urutkan:"), sortSelect), 
		nil, nil, nil, 
		procTable.Table,
	)

	// --- TAB 3: ALERTS & HISTORY ---
	historyList := widget.NewMultiLineEntry()
	historyList.Disable() 
	
	rulesLabel := widget.NewLabel("Aturan Aktif: CPU > 15% (5s)")
	if len(alertEngine.Rules) > 0 {
		r := alertEngine.Rules[0]
		rulesLabel.SetText(fmt.Sprintf("Aturan Aktif: %s > %.0f%% (%ds)", r.Metric, r.Threshold, r.Duration))
	}

	alertTab := container.NewVBox(
		widget.NewLabelWithStyle("Konfigurasi Alert", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		rulesLabel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("History Triggered Alerts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.New(layout.NewGridWrapLayout(fyne.NewSize(500, 200)), historyList),
	)
	procTable.Table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}

		selectedProc := procTable.Data[id.Row-1]

		msg := fmt.Sprintf("Apakah Anda yakin ingin mematikan proses: %s (PID: %d)?", selectedProc.Name, selectedProc.PID)
		confirmDialog := dialog.NewConfirm("Konfirmasi Kill Process", msg, func(ok bool) {
			if ok {
				err := monitor.KillProcess(selectedProc.PID)
				if err != nil {
					dialog.ShowError(err, myWindow)
				} else {
					dialog.ShowInformation("Sukses", "Proses berhasil dihentikan.", myWindow)
				}
			}
			procTable.Table.Unselect(id)
		}, myWindow)

		confirmDialog.Show()
	}

	// --- GABUNGKAN KE TABS ---
	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", dashboardTab),
		container.NewTabItem("Processes", processTab),
		container.NewTabItem("Alert Logs", alertTab),
	)

	myWindow.SetContent(tabs)

	// --- LOOP UTAMA (Goroutine) ---
	go func() {
		for {
			metrics := monitor.GetCurrentMetrics()
			alertEngine.Evaluate(metrics)

			cpuBuffer.Add(metrics.CPUUsage)
			ramBuffer.Add(metrics.RAMUsage)

			cpuLabel.SetText(fmt.Sprintf("CPU: %.2f%%", metrics.CPUUsage))
			ramLabel.SetText(fmt.Sprintf("RAM: %.2f / %.2f GB (%.1f%%)", metrics.RAMUsedGB, metrics.RAMTotalGB, metrics.RAMUsage))
			diskLabel.SetText(fmt.Sprintf("Disk: %.1f%%", metrics.DiskUsage))
			netLabel.SetText(fmt.Sprintf("Net: RX %.1f | TX %.1f KB/s", metrics.NetRXSpeed/1024, metrics.NetTXSpeed/1024))

			sysChart.Update(cpuBuffer.Data, ramBuffer.Data, 500, 150)
			historyList.SetText(strings.Join(alertEngine.History, "\n"))

			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			procTable.UpdateData(monitor.GetTopProcesses())
			time.Sleep(3 * time.Second)
		}
	}()

	myWindow.Resize(fyne.NewSize(550, 650))
	myWindow.ShowAndRun()
}