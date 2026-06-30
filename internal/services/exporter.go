package services

import (
	"encoding/csv"
	"fmt"
	"io"

	"sysmon-app/internal/models"
)

type Exporter struct{}

func NewExporter() *Exporter {
	return &Exporter{}
}

func (e *Exporter) ExportMetrics(writer io.Writer, samples []models.MetricSample) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{"Timestamp", "CPU (%)", "RAM (%)", "Disk (%)", "Net RX (KB/s)", "Net TX (KB/s)"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, s := range samples {
		record := []string{
			s.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", s.CPUUsage),
			fmt.Sprintf("%.2f", s.RAMUsage),
			fmt.Sprintf("%.2f", s.DiskUsage),
			fmt.Sprintf("%.2f", s.NetRXSpeed/1024),
			fmt.Sprintf("%.2f", s.NetTXSpeed/1024),
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}
