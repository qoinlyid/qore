package qore

import "context"

// DependencyStats defines dependency statistical health.
type DependencyStats struct {
	UptimeSeconds     float64 `json:"uptimeSeconds" xml:"uptimeSeconds"`
	UptimeHuman       string  `json:"uptimeHuman" xml:"uptimeHuman"`
	PINGResponse      any     `json:"pingResponse" xml:"pingResponse"`
	PINGLatencyMillis int64   `json:"pingLatencyMillis" xml:"pingLatencyMillis"`
	PINGLatencyHuman  string  `json:"pingLatencyHuman" xml:"pingLatencyHuman"`
	Additional        any     `json:"additional" xml:"additional"`
}

// Dependency is package dependency interface.
type Dependency interface {
	Name() string
	Priority() int
	HealthCheck(ctx context.Context) *DependencyStats
	Open() error
	Close() error
}
