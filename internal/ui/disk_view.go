package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/models"
)


type DiskView struct {
	Container *fyne.Container
	bars      []*widget.ProgressBar
	labels    []*widget.Label
}

func NewDiskView(disks []models.DiskPartition) *DiskView {
	v := &DiskView{Container: container.NewVBox()}

	for range disks {
		bar := widget.NewProgressBar()
		label := widget.NewLabel("")
		v.bars = append(v.bars, bar)
		v.labels = append(v.labels, label)
		v.Container.Add(container.NewVBox(label, bar))
	}

	if len(disks) == 0 {
		v.Container.Add(widget.NewLabel("Tidak ada partisi terdeteksi."))
	}

	v.Update(disks)
	return v
}

func (v *DiskView) Update(disks []models.DiskPartition) {
	for i, d := range disks {
		if i >= len(v.bars) {
			break
		}
		v.labels[i].SetText(fmt.Sprintf("%s  —  %.1f GB / %.1f GB", d.Mountpoint, d.UsedGB, d.TotalGB))
		v.bars[i].SetValue(d.UsedPercent / 100.0)
	}
}
