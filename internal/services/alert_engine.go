package services

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"sysmon-app/internal/models"
)

type AlertEngine struct {
	Rules          []models.AlertRule
	History        []string
	App            fyne.App
	breachCounters map[string]int
	filePath       string
}

func NewAlertEngine(app fyne.App) *AlertEngine {
	e := &AlertEngine{
		Rules:          make([]models.AlertRule, 0),
		History:        make([]string, 0),
		App:            app,
		breachCounters: make(map[string]int),
		filePath:       "alerts.json",
	}
	e.LoadRules()
	return e
}

func (e *AlertEngine) AddRule(rule models.AlertRule) {
	e.Rules = append(e.Rules, rule)
	e.SaveRules()
}

func (e *AlertEngine) SaveRules() {
	data, _ := json.MarshalIndent(e.Rules, "", "  ")
	_ = os.WriteFile(e.filePath, data, 0644)
}

func (e *AlertEngine) LoadRules() {
	data, err := os.ReadFile(e.filePath)
	if err == nil {
		_ = json.Unmarshal(data, &e.Rules)
	}
}

func (e *AlertEngine) Evaluate(metrics models.SystemMetric) {
	for _, rule := range e.Rules {
		var currentValue float64
		
		switch rule.Metric {
		case "CPU":
			currentValue = metrics.CPUUsage
		case "RAM":
			currentValue = metrics.RAMUsage
		}

		ruleKey := fmt.Sprintf("%s-%.1f", rule.Metric, rule.Threshold)

		if currentValue > rule.Threshold {
			e.breachCounters[ruleKey]++ 

			if e.breachCounters[ruleKey] >= rule.Duration {
				timestamp := time.Now().Format("15:04:05")
				msg := fmt.Sprintf("[%s] ALERT: %s melebihi %.1f%%", timestamp, rule.Metric, rule.Threshold)
				
				e.History = append(e.History, msg)

				notification := fyne.NewNotification("Sistem Kritis", msg)
				e.App.SendNotification(notification)

				e.breachCounters[ruleKey] = 0
			}
		} else {
			e.breachCounters[ruleKey] = 0
		}
	}
}