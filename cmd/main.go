package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/models"
	"sysmon-app/internal/repository"
	"sysmon-app/internal/services"
	"sysmon-app/internal/ui"
)

func main() {
	myApp := app.NewWithID("com.sysmon.app")
	myWindow := myApp.NewWindow("SysMON - System Resource Monitor")

	monitor := repository.NewOSMonitor()
	alertEngine := services.NewAlertEngine(myApp)
	cpuBuffer := models.NewRingBuffer(60)
	ramBuffer := models.NewRingBuffer(60)

	cpuLabel := widget.NewLabelWithStyle("- %", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ramLabel := widget.NewLabelWithStyle("- GB", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	diskLabel := widget.NewLabelWithStyle("- %", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	netLabel := widget.NewLabelWithStyle("↓ - | ↑ -", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	metricGrid := container.NewGridWithColumns(4,
		widget.NewCard("CPU", "Processing", cpuLabel),
		widget.NewCard("RAM", "Memory", ramLabel),
		widget.NewCard("Disk", "Storage", diskLabel),
		widget.NewCard("Net (KB/s)", "Bandwidth", netLabel),
	)

	sysChart := ui.NewSystemChart()

	redIndicator := canvas.NewRectangle(color.RGBA{R: 231, G: 76, B: 60, A: 255})
	redIndicator.SetMinSize(fyne.NewSize(12, 12))

	blueIndicator := canvas.NewRectangle(color.RGBA{R: 52, G: 152, B: 219, A: 255})
	blueIndicator.SetMinSize(fyne.NewSize(12, 12))

	legendBox := container.NewHBox(
		layout.NewSpacer(),
		redIndicator, widget.NewLabel("CPU Usage (%)"),
		widget.NewLabel("  |  "),
		blueIndicator, widget.NewLabel("RAM Usage (%)"),
		layout.NewSpacer(),
	)

	chartContent := container.NewVBox(
		legendBox,
		container.New(layout.NewGridWrapLayout(fyne.NewSize(500, 170)), sysChart.Container),
	)
	
	chartCard := widget.NewCard("Live Performance (60s)", "", chartContent)

	exporter := services.NewExporter()
	exportBtn := widget.NewButton("Export CPU History to CSV", func() {
		saveDialog := dialog.NewFileSave(func(w fyne.URIWriteCloser, e error) {
			if w == nil { return }
			defer w.Close() 
			exporter.ExportCPUHistory(w, cpuBuffer.Data)
			dialog.ShowInformation("Sukses", "Data riwayat CPU berhasil diexport!", myWindow)
		}, myWindow)
		
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		saveDialog.SetFileName("cpu_history.csv")
		saveDialog.Show()
	})

	dashboardTab := container.NewPadded(container.NewVBox(
		metricGrid,
		layout.NewSpacer(),
		chartCard,
		layout.NewSpacer(),
		exportBtn,
	))

	procTable := ui.NewProcessTable()
	sortSelect := widget.NewSelect([]string{"CPU", "RAM"}, func(v string) {
		procTable.SortBy = v
		procTable.Table.Refresh()
	})
	sortSelect.SetSelected("CPU")
	
	headerProc := container.NewHBox(
		widget.NewLabelWithStyle("Running Processes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewLabel("Sort By:"),
		sortSelect,
	)

	processTab := container.NewPadded(container.NewBorder(
		container.NewVBox(headerProc, widget.NewSeparator()), 
		nil, nil, nil, 
		procTable.Table,
	))

	historyList := widget.NewMultiLineEntry()
	historyList.Disable() 

	rulesLabel := widget.NewLabel("No active rules.")
	if len(alertEngine.Rules) > 0 {
		r := alertEngine.Rules[0]
		rulesLabel.SetText(fmt.Sprintf("Status: %s > %.0f%% (Duration: %ds)", r.Metric, r.Threshold, r.Duration))
	}

	configCard := widget.NewCard("Konfigurasi Aktif", "Aturan pemicu peringatan", rulesLabel)
	historyCard := widget.NewCard("Log Peringatan", "Riwayat threshold yang terlewati", 
		container.New(layout.NewGridWrapLayout(fyne.NewSize(500, 250)), historyList),
	)

	alertTab := container.NewPadded(container.NewVBox(
		configCard,
		historyCard,
	))

	procTable.Table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 { return }
		selectedProc := procTable.Data[id.Row-1]
		msg := fmt.Sprintf("Hentikan proses %s (PID: %d)?", selectedProc.Name, selectedProc.PID)
		dialog.ShowConfirm("Kill Process", msg, func(ok bool) {
			if ok {
				if err := monitor.KillProcess(selectedProc.PID); err != nil {
					dialog.ShowError(err, myWindow)
				} else {
					dialog.ShowInformation("Sukses", "Proses berhasil dihentikan.", myWindow)
				}
			}
			procTable.Table.Unselect(id)
		}, myWindow)
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", dashboardTab),
		container.NewTabItem("Processes", processTab),
		container.NewTabItem("Alerts", alertTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)

	go func() {
		for {
			metrics := monitor.GetCurrentMetrics()
			alertEngine.Evaluate(metrics)

			cpuBuffer.Add(metrics.CPUUsage)
			ramBuffer.Add(metrics.RAMUsage)

			cpuLabel.SetText(fmt.Sprintf("%.1f %%", metrics.CPUUsage))
			ramLabel.SetText(fmt.Sprintf("%.1f GB\n(%.1f%%)", metrics.RAMUsedGB, metrics.RAMUsage))
			diskLabel.SetText(fmt.Sprintf("%.1f %%", metrics.DiskUsage))
			netLabel.SetText(fmt.Sprintf("↓ %.1f\n↑ %.1f", metrics.NetRXSpeed/1024, metrics.NetTXSpeed/1024))

			sysChart.Update(cpuBuffer.Data, ramBuffer.Data, 500, 170)
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

	myWindow.Resize(fyne.NewSize(600, 500))
	myWindow.ShowAndRun()
}