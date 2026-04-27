package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"sysmon-app/internal/models"
)

// ProcessTable membungkus logika tabel Fyne
type ProcessTable struct {
	Table *widget.Table
	Data  []models.ProcessStat
}

// NewProcessTable adalah Constructor
func NewProcessTable() *ProcessTable {
	pt := &ProcessTable{
		Data: make([]models.ProcessStat, 0),
	}

	// Fyne Table butuh 3 fungsi callback: Ukuran, Template Sel, dan Pengisi Data
	pt.Table = widget.NewTable(
		func() (int, int) {
			return len(pt.Data) + 1, 4 // +1 baris untuk Header, 4 Kolom
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template") // Template dasar sel
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			
			// Jika baris pertama, jadikan Header
			if i.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				switch i.Col {
				case 0: label.SetText("PID")
				case 1: label.SetText("NAMA PROSES")
				case 2: label.SetText("CPU %")
				case 3: label.SetText("RAM %")
				}
				return
			}

			// Isi data proses (aman dari out-of-bounds)
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
	
	// Atur lebar kolom agar rapi
	pt.Table.SetColumnWidth(0, 60)
	pt.Table.SetColumnWidth(1, 200)
	pt.Table.SetColumnWidth(2, 70)
	pt.Table.SetColumnWidth(3, 70)

	return pt // INI YANG SEBELUMNYA HILANG
}

// UpdateData memperbarui isi tabel (INI JUGA SEBELUMNYA HILANG)
func (pt *ProcessTable) UpdateData(newData []models.ProcessStat) {
	pt.Data = newData
	pt.Table.Refresh()
}