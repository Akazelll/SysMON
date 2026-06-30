package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
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

	// Riwayat metrik berstempel waktu untuk export (cap ~1800 sampel ≈ 1 jam).
	metricHistory := models.NewMetricHistory(1800)

	// Pembacaan awal (blocking ~1 dtk) untuk menentukan jumlah core & partisi
	initialMetrics := monitor.GetCurrentMetrics()

	// ===================== METRIC CARDS =====================
	cpuLabel := widget.NewLabelWithStyle("- %", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ramLabel := widget.NewLabelWithStyle("- GB", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	diskLabel := widget.NewLabelWithStyle("- %", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	netLabel := widget.NewLabelWithStyle("↓ - | ↑ -", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	metricGrid := container.NewGridWithColumns(4,
		widget.NewCard("CPU", "Total", cpuLabel),
		widget.NewCard("RAM", "Memory", ramLabel),
		widget.NewCard("Disk", "Primary", diskLabel),
		widget.NewCard("Net (KB/s)", "Bandwidth", netLabel),
	)

	// ===================== CPU PER-CORE =====================
	coreView := ui.NewCPUCoreView(len(initialMetrics.CPUPerCore))
	coreCard := widget.NewCard("CPU Per-Core", "Beban tiap logical core", coreView.Container)

	// ===================== LIVE CHART =====================
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

	// ===================== DISK PARTITIONS =====================
	diskView := ui.NewDiskView(initialMetrics.Disks)
	diskCard := widget.NewCard("Disk Partitions", "Penggunaan tiap partisi", diskView.Container)

	// ===================== EXPORT =====================
	exporter := services.NewExporter()
	rangeSelect := widget.NewSelect([]string{"1 Menit", "5 Menit", "15 Menit", "Semua"}, nil)
	rangeSelect.SetSelected("Semua")

	exportBtn := widget.NewButton("Export ke CSV", func() {
		var samples []models.MetricSample
		switch rangeSelect.Selected {
		case "1 Menit":
			samples = metricHistory.Since(1 * time.Minute)
		case "5 Menit":
			samples = metricHistory.Since(5 * time.Minute)
		case "15 Menit":
			samples = metricHistory.Since(15 * time.Minute)
		default:
			samples = metricHistory.All()
		}

		if len(samples) == 0 {
			dialog.ShowInformation("Info", "Belum ada data pada rentang tersebut.", myWindow)
			return
		}

		saveDialog := dialog.NewFileSave(func(w fyne.URIWriteCloser, e error) {
			if e != nil || w == nil {
				return
			}
			defer w.Close()
			if err := exporter.ExportMetrics(w, samples); err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			dialog.ShowInformation("Sukses", fmt.Sprintf("%d sampel metrik berhasil diexport!", len(samples)), myWindow)
		}, myWindow)

		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		saveDialog.SetFileName("metric_history.csv")
		saveDialog.Show()
	})

	exportCard := widget.NewCard("Export Data", "Pilih rentang waktu lalu simpan riwayat metrik ke CSV",
		container.NewHBox(widget.NewLabel("Range Waktu:"), rangeSelect, layout.NewSpacer(), exportBtn),
	)

	dashboardTab := container.NewPadded(container.NewVScroll(container.NewVBox(
		metricGrid,
		coreCard,
		chartCard,
		diskCard,
		exportCard,
	)))

	// ===================== PROCESS LIST =====================
	procTable := ui.NewProcessTable()
	sortSelect := widget.NewSelect([]string{"CPU", "RAM"}, func(v string) {
		procTable.SortBy = v
		procTable.UpdateData(procTable.Data)
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

	// ===================== ALERTS TAB =====================
	metricSelect := widget.NewSelect([]string{"CPU", "RAM"}, nil)
	metricSelect.SetSelected("CPU")
	thresholdEntry := widget.NewEntry()
	thresholdEntry.SetPlaceHolder("contoh: 90")
	durationEntry := widget.NewEntry()
	durationEntry.SetPlaceHolder("contoh: 30")

	// Daftar aturan aktif + tombol hapus, dibangun ulang tiap ada perubahan.
	rulesBox := container.NewVBox()
	var refreshRules func()
	refreshRules = func() {
		rulesBox.RemoveAll()
		rules := alertEngine.GetRules()
		if len(rules) == 0 {
			rulesBox.Add(widget.NewLabel("Belum ada aturan aktif."))
		}
		for i, r := range rules {
			idx := i
			text := fmt.Sprintf("%s  >  %.0f%%  selama  %d detik", r.Metric, r.Threshold, r.Duration)
			delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				alertEngine.DeleteRule(idx)
				refreshRules()
			})
			delBtn.Importance = widget.DangerImportance
			rulesBox.Add(container.NewBorder(nil, nil, nil, delBtn, widget.NewLabel(text)))
		}
		rulesBox.Refresh()
	}
	refreshRules()

	addForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Metrik", Widget: metricSelect},
			{Text: "Threshold (%)", Widget: thresholdEntry},
			{Text: "Durasi (detik)", Widget: durationEntry},
		},
		SubmitText: "Tambah Aturan",
		OnSubmit: func() {
			threshold, errT := strconv.ParseFloat(strings.TrimSpace(thresholdEntry.Text), 64)
			if errT != nil || threshold <= 0 || threshold > 100 {
				dialog.ShowError(fmt.Errorf("threshold harus angka antara 1 - 100"), myWindow)
				return
			}
			duration, errD := strconv.Atoi(strings.TrimSpace(durationEntry.Text))
			if errD != nil || duration < 0 {
				dialog.ShowError(fmt.Errorf("durasi harus angka detik (>= 0)"), myWindow)
				return
			}
			alertEngine.AddRule(models.AlertRule{
				Metric:    metricSelect.Selected,
				Threshold: threshold,
				Duration:  duration,
			})
			thresholdEntry.SetText("")
			durationEntry.SetText("")
			refreshRules()
			dialog.ShowInformation("Sukses", "Aturan alert ditambahkan & disimpan.", myWindow)
		},
	}

	historyList := widget.NewMultiLineEntry()
	historyList.Disable()

	formCard := widget.NewCard("Buat Aturan Alert", "Picu notifikasi saat ambang terlampaui", addForm)
	rulesCard := widget.NewCard("Aturan Aktif", "Tersimpan di alerts.json (aktif setelah restart)", rulesBox)
	historyCard := widget.NewCard("Log Peringatan", "Riwayat alert ter-trigger",
		container.New(layout.NewGridWrapLayout(fyne.NewSize(500, 200)), historyList),
	)

	alertTab := container.NewPadded(container.NewVScroll(container.NewVBox(
		formCard,
		rulesCard,
		historyCard,
	)))

	// ===================== TABS =====================
	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", dashboardTab),
		container.NewTabItem("Processes", processTab),
		container.NewTabItem("Alerts", alertTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)

	// ===================== SAMPLING LOOPS =====================
	go func() {
		for {
			metrics := monitor.GetCurrentMetrics()
			alertEngine.Evaluate(metrics)

			cpuBuffer.Add(metrics.CPUUsage)
			ramBuffer.Add(metrics.RAMUsage)
			metricHistory.Add(models.MetricSample{
				Time:       time.Now(),
				CPUUsage:   metrics.CPUUsage,
				RAMUsage:   metrics.RAMUsage,
				DiskUsage:  metrics.DiskUsage,
				NetRXSpeed: metrics.NetRXSpeed,
				NetTXSpeed: metrics.NetTXSpeed,
			})

			cpuLabel.SetText(fmt.Sprintf("%.1f %%", metrics.CPUUsage))
			ramLabel.SetText(fmt.Sprintf("%.1f GB\n(%.1f%%)", metrics.RAMUsedGB, metrics.RAMUsage))
			diskLabel.SetText(fmt.Sprintf("%.1f %%", metrics.DiskUsage))
			netLabel.SetText(fmt.Sprintf("↓ %.1f\n↑ %.1f", metrics.NetRXSpeed/1024, metrics.NetTXSpeed/1024))

			coreView.Update(metrics.CPUPerCore)
			diskView.Update(metrics.Disks)
			sysChart.Update(cpuBuffer.Data, ramBuffer.Data, 500, 170)
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

	myWindow.Resize(fyne.NewSize(720, 680))
	myWindow.ShowAndRun()
}
