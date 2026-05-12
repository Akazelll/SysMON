package models

type AlertRule struct {
	Metric    string
	Threshold float64
	Duration  int
}