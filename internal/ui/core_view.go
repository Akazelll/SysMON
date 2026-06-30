package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)


type CPUCoreView struct {
	Container *fyne.Container
	bars      []*widget.ProgressBar
}

func NewCPUCoreView(numCores int) *CPUCoreView {
	v := &CPUCoreView{}
	cells := make([]fyne.CanvasObject, 0, numCores)

	for i := 0; i < numCores; i++ {
		bar := widget.NewProgressBar()
		v.bars = append(v.bars, bar)
		label := widget.NewLabel(fmt.Sprintf("Core %d", i))
		cells = append(cells, container.NewBorder(nil, nil, label, nil, bar))
	}

	v.Container = container.NewGridWithColumns(2, cells...)
	return v
}

func (v *CPUCoreView) Update(perCore []float64) {
	for i, val := range perCore {
		if i >= len(v.bars) {
			break
		}
		v.bars[i].SetValue(val / 100.0)
	}
}
