package services

import (
	"encoding/csv"
	"fmt"
	"io"
)

type Exporter struct{}

func NewExporter() *Exporter {
	return &Exporter{}
}

func (e *Exporter) ExportCPUHistory(writer io.Writer, cpuData []float64) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	_ = csvWriter.Write([]string{"Detik Ke (Mundur)", "CPU Usage (%)"})

	for i, val := range cpuData {
		record := []string{
			fmt.Sprintf("%d", len(cpuData)-i),
			fmt.Sprintf("%.2f", val),
		}
		_ = csvWriter.Write(record)
	}

	return nil
}