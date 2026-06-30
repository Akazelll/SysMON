package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"sysmon-app/internal/models"
)

const maxHistoryEntries = 200

type AlertEngine struct {
	Rules       []models.AlertRule
	History     []string
	App         fyne.App
	breachSince map[string]time.Time
	rulesPath   string
	historyPath string
	mu          sync.Mutex 
}

func NewAlertEngine(app fyne.App) *AlertEngine {
	e := &AlertEngine{
		Rules:       make([]models.AlertRule, 0),
		History:     make([]string, 0),
		App:         app,
		breachSince: make(map[string]time.Time),
		rulesPath:   "alerts.json",
		historyPath: "alert_history.json",
	}
	e.load() 
	return e
}



func (e *AlertEngine) load() {
	if data, err := os.ReadFile(e.rulesPath); err == nil {
		_ = json.Unmarshal(data, &e.Rules)
	}
	if data, err := os.ReadFile(e.historyPath); err == nil {
		_ = json.Unmarshal(data, &e.History)
	}
}

func (e *AlertEngine) saveRulesLocked() {
	if data, err := json.MarshalIndent(e.Rules, "", "  "); err == nil {
		_ = os.WriteFile(e.rulesPath, data, 0644)
	}
}

func (e *AlertEngine) saveHistoryLocked() {
	if data, err := json.MarshalIndent(e.History, "", "  "); err == nil {
		_ = os.WriteFile(e.historyPath, data, 0644)
	}
}


func (e *AlertEngine) AddRule(rule models.AlertRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Rules = append(e.Rules, rule)
	e.saveRulesLocked()
}

func (e *AlertEngine) DeleteRule(index int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if index < 0 || index >= len(e.Rules) {
		return
	}
	e.Rules = append(e.Rules[:index], e.Rules[index+1:]...)
	e.saveRulesLocked()
}

func (e *AlertEngine) GetRules() []models.AlertRule {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]models.AlertRule, len(e.Rules))
	copy(out, e.Rules)
	return out
}

func (e *AlertEngine) GetHistory() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]string, len(e.History))
	copy(out, e.History)
	return out
}

func ruleKey(r models.AlertRule) string {
	return fmt.Sprintf("%s-%.1f-%d", r.Metric, r.Threshold, r.Duration)
}

func (e *AlertEngine) Evaluate(metrics models.SystemMetric) {
	e.mu.Lock()

	now := time.Now()
	var triggered []string

	for _, rule := range e.Rules {
		var currentValue float64
		switch rule.Metric {
		case "CPU":
			currentValue = metrics.CPUUsage
		case "RAM":
			currentValue = metrics.RAMUsage
		default:
			continue
		}

		key := ruleKey(rule)

		if currentValue <= rule.Threshold {
			delete(e.breachSince, key)
			continue
		}

		start, ok := e.breachSince[key]
		if !ok {
			e.breachSince[key] = now
			continue
		}

		if now.Sub(start) >= time.Duration(rule.Duration)*time.Second {
			timestamp := now.Format("2006-01-02 15:04:05")
			msg := fmt.Sprintf("[%s] ALERT: %s melebihi %.0f%% selama %d detik (nilai: %.1f%%)",
				timestamp, rule.Metric, rule.Threshold, rule.Duration, currentValue)

			e.History = append(e.History, msg)
			if len(e.History) > maxHistoryEntries {
				e.History = e.History[len(e.History)-maxHistoryEntries:]
			}
			triggered = append(triggered, msg)

			e.breachSince[key] = now
		}
	}

	if len(triggered) > 0 {
		e.saveHistoryLocked()
	}
	e.mu.Unlock()

	for _, msg := range triggered {
		e.App.SendNotification(fyne.NewNotification("Sistem Kritis", msg))
	}
}
