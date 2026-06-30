package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/ui"
)


func buildProcessTab() (fyne.CanvasObject, *ui.ProcessTable) {
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

	return processTab, procTable
}
