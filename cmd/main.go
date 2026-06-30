package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"sysmon-app/internal/models"
	"sysmon-app/internal/repository"
	"sysmon-app/internal/services"
)

func main() {
	myApp := app.NewWithID("com.sysmon.app")
	myWindow := myApp.NewWindow("SysMON - System Resource Monitor")

	// ===================== DEPENDENCIES =====================
	monitor := repository.NewOSMonitor()
	alertEngine := services.NewAlertEngine(myApp)
	cpuBuffer := models.NewRingBuffer(60)
	ramBuffer := models.NewRingBuffer(60)

	// Riwayat metrik berstempel waktu untuk export (cap ~1800 sampel ≈ 1 jam).
	metricHistory := models.NewMetricHistory(1800)

	// Pembacaan awal (blocking ~1 dtk) untuk menentukan jumlah core & partisi.
	initialMetrics := monitor.GetCurrentMetrics()

	// ===================== TABS =====================
	dashboardTab, dashUI := buildDashboardTab(myWindow, initialMetrics, metricHistory)
	processTab, procTable := buildProcessTab()
	alertTab, historyList := buildAlertTab(myWindow, alertEngine)

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", dashboardTab),
		container.NewTabItem("Processes", processTab),
		container.NewTabItem("Alerts", alertTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	myWindow.SetContent(tabs)

	// ===================== SAMPLING LOOPS =====================
	startSampling(monitor, alertEngine, cpuBuffer, ramBuffer, metricHistory, dashUI, historyList, procTable)

	myWindow.Resize(fyne.NewSize(720, 680))
	myWindow.ShowAndRun()
}
