package models

import (
	"sync"
	"time"
)

type MetricSample struct {
	Time       time.Time
	CPUUsage   float64
	RAMUsage   float64
	DiskUsage  float64
	NetRXSpeed float64 
	NetTXSpeed float64 
}

type MetricHistory struct {
	samples []MetricSample
	maxSize int
	mu      sync.Mutex
}

func NewMetricHistory(maxSize int) *MetricHistory {
	return &MetricHistory{
		samples: make([]MetricSample, 0, maxSize),
		maxSize: maxSize,
	}
}

func (h *MetricHistory) Add(s MetricSample) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.samples) >= h.maxSize {
		h.samples = h.samples[1:]
	}
	h.samples = append(h.samples, s)
}

func (h *MetricHistory) All() []MetricSample {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]MetricSample, len(h.samples))
	copy(out, h.samples)
	return out
}

func (h *MetricHistory) Since(d time.Duration) []MetricSample {
	h.mu.Lock()
	defer h.mu.Unlock()
	cutoff := time.Now().Add(-d)
	out := make([]MetricSample, 0)
	for _, s := range h.samples {
		if s.Time.After(cutoff) {
			out = append(out, s)
		}
	}
	return out
}
