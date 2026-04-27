package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// SystemChart membungkus logika penggambaran grafik multi-garis
type SystemChart struct {
	Container *fyne.Container
}

// NewSystemChart adalah Constructor
func NewSystemChart() *SystemChart {
	return &SystemChart{
		Container: container.NewWithoutLayout(),
	}
}

// Update menggambar ulang grid, garis CPU, dan garis RAM
func (sc *SystemChart) Update(cpuData []float64, ramData []float64, width float32, height float32) {
	sc.Container.RemoveAll() // Bersihkan frame sebelumnya

	// 1. Gambar Background Grid (Garis bantu horizontal)
	gridColor := color.RGBA{R: 150, G: 150, B: 150, A: 80} // Abu-abu transparan
	for i := 1; i <= 4; i++ {
		y := height - (float32(i) * 25.0 / 100.0 * height)
		
		gridLine := canvas.NewLine(gridColor)
		gridLine.StrokeWidth = 1
		gridLine.Position1 = fyne.NewPos(0, y)
		gridLine.Position2 = fyne.NewPos(width, y)
		sc.Container.Add(gridLine)
	}

	// Fungsi internal (Helper) untuk menggambar garis metrik
	drawLine := func(data []float64, lineColor color.Color) {
		if len(data) < 2 {
			return
		}
		stepX := width / float32(len(data)-1)
		for i := 0; i < len(data)-1; i++ {
			x1 := float32(i) * stepX
			y1 := height - (float32(data[i]) / 100.0 * height)
			
			x2 := float32(i+1) * stepX
			y2 := height - (float32(data[i+1]) / 100.0 * height)

			line := canvas.NewLine(lineColor)
			line.StrokeWidth = 2
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			
			sc.Container.Add(line)
		}
	}

	// 2. Gambar Garis CPU (Merah Terang)
	drawLine(cpuData, color.RGBA{R: 231, G: 76, B: 60, A: 255})
	
	// 3. Gambar Garis RAM (Biru Terang)
	drawLine(ramData, color.RGBA{R: 52, G: 152, B: 219, A: 255})

	// Render ulang ke layar
	sc.Container.Refresh()
}