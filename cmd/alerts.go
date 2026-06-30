package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/models"
	"sysmon-app/internal/services"
)

func buildAlertTab(win fyne.Window, alertEngine *services.AlertEngine) (fyne.CanvasObject, *widget.Entry) {
	metricSelect := widget.NewSelect([]string{"CPU", "RAM"}, nil)
	metricSelect.SetSelected("CPU")
	thresholdEntry := widget.NewEntry()
	thresholdEntry.SetPlaceHolder("contoh: 90")
	durationEntry := widget.NewEntry()
	durationEntry.SetPlaceHolder("contoh: 30")

	
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
				dialog.ShowError(fmt.Errorf("threshold harus angka antara 1 - 100"), win)
				return
			}
			duration, errD := strconv.Atoi(strings.TrimSpace(durationEntry.Text))
			if errD != nil || duration < 0 {
				dialog.ShowError(fmt.Errorf("durasi harus angka detik (>= 0)"), win)
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
			dialog.ShowInformation("Sukses", "Aturan alert ditambahkan & disimpan.", win)
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

	return alertTab, historyList
}
