package main

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/models"
	"sysmon-app/internal/services"
	"sysmon-app/internal/ui"
)

type dashboardUI struct {
	cpuLabel  *widget.Label
	ramLabel  *widget.Label
	diskLabel *widget.Label
	netLabel  *widget.Label
	coreView  *ui.CPUCoreView
	diskView  *ui.DiskView
	sysChart  *ui.SystemChart
}

func buildDashboardTab(win fyne.Window, initial models.SystemMetric, history *models.MetricHistory) (fyne.CanvasObject, *dashboardUI) {
	d := &dashboardUI{}

	
	d.cpuLabel = widget.NewLabelWithStyle("- %", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	d.ramLabel = widget.NewLabelWithStyle("- GB", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	d.diskLabel = widget.NewLabelWithStyle("- %", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	d.netLabel = widget.NewLabelWithStyle("↓ - | ↑ -", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	metricGrid := container.NewGridWithColumns(4,
		widget.NewCard("CPU", "Total", d.cpuLabel),
		widget.NewCard("RAM", "Memory", d.ramLabel),
		widget.NewCard("Disk", "Primary", d.diskLabel),
		widget.NewCard("Net (KB/s)", "Bandwidth", d.netLabel),
	)

	d.coreView = ui.NewCPUCoreView(len(initial.CPUPerCore))
	coreCard := widget.NewCard("CPU Per-Core", "Beban tiap logical core", d.coreView.Container)

	d.sysChart = ui.NewSystemChart()

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
		container.New(layout.NewGridWrapLayout(fyne.NewSize(500, 170)), d.sysChart.Container),
	)
	chartCard := widget.NewCard("Live Performance (60s)", "", chartContent)

	d.diskView = ui.NewDiskView(initial.Disks)
	diskCard := widget.NewCard("Disk Partitions", "Penggunaan tiap partisi", d.diskView.Container)

	exportCard := buildExportCard(win, history)

	dashboardTab := container.NewPadded(container.NewVScroll(container.NewVBox(
		metricGrid,
		coreCard,
		chartCard,
		diskCard,
		exportCard,
	)))

	return dashboardTab, d
}

func buildExportCard(win fyne.Window, history *models.MetricHistory) fyne.CanvasObject {
	exporter := services.NewExporter()
	rangeSelect := widget.NewSelect([]string{"1 Menit", "5 Menit", "15 Menit", "Semua"}, nil)
	rangeSelect.SetSelected("Semua")

	exportBtn := widget.NewButton("Export ke CSV", func() {
		var samples []models.MetricSample
		switch rangeSelect.Selected {
		case "1 Menit":
			samples = history.Since(1 * time.Minute)
		case "5 Menit":
			samples = history.Since(5 * time.Minute)
		case "15 Menit":
			samples = history.Since(15 * time.Minute)
		default:
			samples = history.All()
		}

		if len(samples) == 0 {
			dialog.ShowInformation("Info", "Belum ada data pada rentang tersebut.", win)
			return
		}

		saveDialog := dialog.NewFileSave(func(w fyne.URIWriteCloser, e error) {
			if e != nil || w == nil {
				return
			}
			defer w.Close()
			if err := exporter.ExportMetrics(w, samples); err != nil {
				dialog.ShowError(err, win)
				return
			}
			dialog.ShowInformation("Sukses", fmt.Sprintf("%d sampel metrik berhasil diexport!", len(samples)), win)
		}, win)

		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		saveDialog.SetFileName("metric_history.csv")
		saveDialog.Show()
	})

	return widget.NewCard("Export Data", "Pilih rentang waktu lalu simpan riwayat metrik ke CSV",
		container.NewHBox(widget.NewLabel("Range Waktu:"), rangeSelect, layout.NewSpacer(), exportBtn),
	)
}
