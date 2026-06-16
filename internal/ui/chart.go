package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

type SystemChart struct {
	Container *fyne.Container
}

func NewSystemChart() *SystemChart {
	return &SystemChart{
		Container: container.NewWithoutLayout(),
	}
}

func (sc *SystemChart) Update(cpuData []float64, ramData []float64, width float32, height float32) {
	sc.Container.RemoveAll()
	
	gridColor := theme.DisabledColor() 
	textColor := theme.ForegroundColor()
	
	leftPadding := float32(35.0)
	chartWidth := width - leftPadding

	for i := 0; i <= 4; i++ {
		percent := float32(i) * 25.0
		y := height - (percent / 100.0 * height)
		
		lbl := canvas.NewText(fmt.Sprintf("%.0f%%", percent), textColor)
		lbl.TextSize = 10
		lbl.Move(fyne.NewPos(0, y-6))
		sc.Container.Add(lbl)

		gridLine := canvas.NewLine(gridColor)
		gridLine.StrokeWidth = 1
		gridLine.Position1 = fyne.NewPos(leftPadding, y)
		gridLine.Position2 = fyne.NewPos(width, y)
		sc.Container.Add(gridLine)
	}

	drawLine := func(data []float64, lineColor color.Color) {
		if len(data) < 2 {
			return
		}
		stepX := chartWidth / float32(len(data)-1)
		
		for i := 0; i < len(data)-1; i++ {
			x1 := leftPadding + float32(i)*stepX
			y1 := height - (float32(data[i]) / 100.0 * height)
			
			x2 := leftPadding + float32(i+1)*stepX
			y2 := height - (float32(data[i+1]) / 100.0 * height)

			line := canvas.NewLine(lineColor)
			line.StrokeWidth = 2.5 
			line.Position1 = fyne.NewPos(x1, y1)
			line.Position2 = fyne.NewPos(x2, y2)
			
			sc.Container.Add(line)
		}
	}
	
	drawLine(cpuData, color.RGBA{R: 231, G: 76, B: 60, A: 255})
	drawLine(ramData, color.RGBA{R: 52, G: 152, B: 219, A: 255})

	sc.Container.Refresh()
}