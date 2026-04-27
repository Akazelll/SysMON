package services

import (
	"encoding/csv"
	"fmt"
	"io"
)

// Exporter bertugas menangani logika konversi data ke format file
type Exporter struct{}

// NewExporter adalah Constructor
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportCPUHistory mengubah array data CPU menjadi baris-baris CSV
func (e *Exporter) ExportCPUHistory(writer io.Writer, cpuData []float64) error {
	csvWriter := csv.NewWriter(writer)
	// Pastikan data disiram (flush) ke file saat fungsi selesai
	defer csvWriter.Flush()

	// 1. Tulis Header (Baris pertama di Excel)
	_ = csvWriter.Write([]string{"Detik Ke (Mundur)", "CPU Usage (%)"})

	// 2. Tulis Data (Looping dari data terbaru ke terlama)
	for i, val := range cpuData {
		record := []string{
			fmt.Sprintf("%d", len(cpuData)-i), // Anggap indeks terakhir adalah detik ini
			fmt.Sprintf("%.2f", val),
		}
		_ = csvWriter.Write(record)
	}

	return nil
}