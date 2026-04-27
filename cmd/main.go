package main

import (
	"fmt"
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
	myWindow := myApp.NewWindow("System Resource Monitor")

	cpuLabel := widget.NewLabel("CPU: - %")
	ramLabel := widget.NewLabel("RAM: - %")
	diskLabel := widget.NewLabel("Disk: - %")
	netLabel := widget.NewLabel("Net: RX - KB/s | TX - KB/s")

	// 1. Siapkan DUA Buffer (CPU dan RAM)
	cpuBuffer := models.NewRingBuffer(60)
	ramBuffer := models.NewRingBuffer(60)

	// 2. Inisialisasi Chart Baru
	sysChart := ui.NewSystemChart()
	chartContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(400, 120)), sysChart.Container)

	// 3. Tambahkan Legend (Keterangan Warna) di atas chart
	legend := widget.NewLabel("Grafik 60 Detik Terakhir: 🟥 CPU | 🟦 RAM")
	legend.Alignment = fyne.TextAlignCenter

	// Tombol Export
	exporter := services.NewExporter()
	exportBtn := widget.NewButton("Export CPU History to CSV", func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()
			exporter.ExportCPUHistory(writer, cpuBuffer.Data)
			dialog.ShowInformation("Sukses", "Data riwayat CPU berhasil diexport!", myWindow)
		}, myWindow)
	})

	// Susun bagian atas
	topSection := container.NewVBox(
		exportBtn,
		cpuLabel,
		ramLabel,
		diskLabel,
		netLabel,
		legend,          // Legend masuk ke UI
		chartContainer,  // Chart sekarang ada di bawah legend
	)

	procTable := ui.NewProcessTable()
	split := container.NewVSplit(topSection, procTable.Table)
	split.Offset = 0.55 // Kasih ruang lebih sedikit untuk metrik atas agar tabel lega

	myWindow.SetContent(split)

	monitor := repository.NewOSMonitor()
	alertEngine := services.NewAlertEngine(myApp)

	alertEngine.AddRule(models.AlertRule{
		Metric:    "CPU",
		Threshold: 15.0,
		Duration:  5,
	})

	// Loop Metrik (1 Detik)
	go func() {
		for {
			metrics := monitor.GetCurrentMetrics()
			alertEngine.Evaluate(metrics)

			// Masukkan data ke dua buffer
			cpuBuffer.Add(metrics.CPUUsage)
			ramBuffer.Add(metrics.RAMUsage)

			rxKB := metrics.NetRXSpeed / 1024
			txKB := metrics.NetTXSpeed / 1024

			cpuLabel.SetText(fmt.Sprintf("CPU Usage: %.2f%%", metrics.CPUUsage))
			ramLabel.SetText(fmt.Sprintf("RAM Usage: %.2f%%", metrics.RAMUsage))
			diskLabel.SetText(fmt.Sprintf("Disk Usage: %.2f%%", metrics.DiskUsage))
			netLabel.SetText(fmt.Sprintf("Network: RX %.2f KB/s | TX %.2f KB/s", rxKB, txKB))

			// Update Chart dengan dua dataset sekaligus
			sysChart.Update(cpuBuffer.Data, ramBuffer.Data, 400, 120)

			time.Sleep(1 * time.Second)
		}
	}()

	// Loop Tabel (3 Detik)
	go func() {
		for {
			topProcs := monitor.GetTopProcesses()
			procTable.UpdateData(topProcs)
			time.Sleep(3 * time.Second)
		}
	}()

	myWindow.Resize(fyne.NewSize(500, 650))
	myWindow.ShowAndRun()
}