package ui

import (
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/models"
)

type ProcessTable struct {
	Table  *widget.Table
	Data   []models.ProcessStat
	SortBy string 
}

func NewProcessTable() *ProcessTable {
	pt := &ProcessTable{
		Data:   make([]models.ProcessStat, 0),
		SortBy: "CPU",
	}

	pt.Table = widget.NewTable(
		func() (int, int) {
			return len(pt.Data) + 1, 4
		},
		func() fyne.CanvasObject {
			// Center text layout for better readability in cells
			lbl := widget.NewLabel("")
			lbl.Alignment = fyne.TextAlignCenter
			return lbl
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			
			if i.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				switch i.Col {
				case 0: label.SetText("PID")
				case 1: 
					label.SetText("PROCESS NAME")
					label.Alignment = fyne.TextAlignLeading
				case 2: 
					txt := "CPU %"
					if pt.SortBy == "CPU" { txt += " ↓" }
					label.SetText(txt)
				case 3: 
					txt := "RAM %"
					if pt.SortBy == "RAM" { txt += " ↓" }
					label.SetText(txt)
				}
				return
			}

			if i.Row-1 < len(pt.Data) {
				proc := pt.Data[i.Row-1]
				label.TextStyle = fyne.TextStyle{Bold: false}
				switch i.Col {
				case 0: label.SetText(fmt.Sprintf("%d", proc.PID))
				case 1: 
					label.SetText(proc.Name)
					label.Alignment = fyne.TextAlignLeading
				case 2: label.SetText(fmt.Sprintf("%.1f", proc.CPUUsage))
				case 3: label.SetText(fmt.Sprintf("%.1f", proc.RAMUsage))
				}
			}
		},
	)

	// Proporsi lebar disesuaikan dengan ukuran window yang baru
	pt.Table.SetColumnWidth(0, 60)
	pt.Table.SetColumnWidth(1, 250) // Ruang lebih besar untuk nama proses
	pt.Table.SetColumnWidth(2, 90)
	pt.Table.SetColumnWidth(3, 90)

	return pt
}

func (pt *ProcessTable) sortData() {
	sort.Slice(pt.Data, func(i, j int) bool {
		if pt.SortBy == "RAM" {
			return pt.Data[i].RAMUsage > pt.Data[j].RAMUsage
		}
		return pt.Data[i].CPUUsage > pt.Data[j].CPUUsage
	})
}

func (pt *ProcessTable) UpdateData(newData []models.ProcessStat) {
	pt.Data = newData
	pt.sortData()
	pt.Table.Refresh()
}