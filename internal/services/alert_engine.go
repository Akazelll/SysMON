package services

import (
	"fmt"

	"fyne.io/fyne/v2"

	"sysmon-app/internal/models"
)

// AlertEngine bertugas mengevaluasi metrik terhadap aturan yang ada
type AlertEngine struct {
	Rules          []models.AlertRule
	App            fyne.App
	breachCounters map[string]int // Menghitung berapa detik batas terlampaui
}

// NewAlertEngine adalah Constructor
func NewAlertEngine(app fyne.App) *AlertEngine {
	return &AlertEngine{
		Rules:          make([]models.AlertRule, 0),
		App:            app,
		breachCounters: make(map[string]int),
	}
}

// AddRule menambahkan aturan baru ke dalam mesin
func (e *AlertEngine) AddRule(rule models.AlertRule) {
	e.Rules = append(e.Rules, rule)
}

// Evaluate dipanggil setiap 1 detik untuk mengecek kondisi metrik terbaru
func (e *AlertEngine) Evaluate(metrics models.SystemMetric) {
	for _, rule := range e.Rules {
		var currentValue float64
		if rule.Metric == "CPU" {
			currentValue = metrics.CPUUsage
		} else if rule.Metric == "RAM" {
			currentValue = metrics.RAMUsage
		}

		// Buat kunci unik untuk tiap aturan
		ruleKey := fmt.Sprintf("%s-%.1f", rule.Metric, rule.Threshold)

		// Evaluasi
		if currentValue > rule.Threshold {
			e.breachCounters[ruleKey]++ // Tambah 1 detik

			// Jika durasi terlampaui
			if e.breachCounters[ruleKey] >= rule.Duration {
				// Kirim notifikasi bawaan OS!
				title := "Peringatan Sistem: " + rule.Metric
				content := fmt.Sprintf("%s melebihi batas %.1f%% selama %d detik!", rule.Metric, rule.Threshold, rule.Duration)
				
				notification := fyne.NewNotification(title, content)
				e.App.SendNotification(notification)

				// Reset counter agar tidak spam tiap detik setelah notifikasi muncul
				e.breachCounters[ruleKey] = 0
			}
		} else {
			// Jika metrik turun kembali sebelum batas waktu, reset hitungan ke 0
			e.breachCounters[ruleKey] = 0
		}
	}
}