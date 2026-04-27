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
	SortBy string // Menyimpan kolom mana yang sedang di-sort ("CPU" atau "RAM")
}

func NewProcessTable() *ProcessTable {
	pt := &ProcessTable{
		Data:   make([]models.ProcessStat, 0),
		SortBy: "CPU", // Default sort berdasarkan CPU
	}

	pt.Table = widget.NewTable(
		func() (int, int) {
			return len(pt.Data) + 1, 4
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			
			// --- LOGIKA HEADER ---
			if i.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				switch i.Col {
				case 0: label.SetText("PID")
				case 1: label.SetText("NAMA PROSES")
				case 2: 
					txt := "CPU %"
					if pt.SortBy == "CPU" { txt += " ↓" } // Beri tanda jika sedang aktif
					label.SetText(txt)
				case 3: 
					txt := "RAM %"
					if pt.SortBy == "RAM" { txt += " ↓" }
					label.SetText(txt)
				}
				return
			}

			// --- LOGIKA DATA ---
			if i.Row-1 < len(pt.Data) {
				proc := pt.Data[i.Row-1]
				label.TextStyle = fyne.TextStyle{Bold: false}
				switch i.Col {
				case 0: label.SetText(fmt.Sprintf("%d", proc.PID))
				case 1: label.SetText(proc.Name)
				case 2: label.SetText(fmt.Sprintf("%.1f", proc.CPUUsage))
				case 3: label.SetText(fmt.Sprintf("%.1f", proc.RAMUsage))
				}
			}
		},
	)

	pt.Table.SetColumnWidth(0, 60)
	pt.Table.SetColumnWidth(1, 200)
	pt.Table.SetColumnWidth(2, 75)
	pt.Table.SetColumnWidth(3, 75)

	return pt
}

// sortData adalah fungsi internal untuk mengurutkan array Data
func (pt *ProcessTable) sortData() {
	sort.Slice(pt.Data, func(i, j int) bool {
		if pt.SortBy == "RAM" {
			return pt.Data[i].RAMUsage > pt.Data[j].RAMUsage
		}
		return pt.Data[i].CPUUsage > pt.Data[j].CPUUsage
	})
}

// UpdateData sekarang otomatis melakukan sorting sebelum refresh tampilan
func (pt *ProcessTable) UpdateData(newData []models.ProcessStat) {
	pt.Data = newData
	pt.sortData() // Urutkan dulu
	pt.Table.Refresh()
}