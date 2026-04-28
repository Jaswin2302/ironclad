package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAlertStateNoAlertBelowThreshold(t *testing.T) {
	alerts := &AlertState{}
	metrics := Metrics{
		Hostname:   "test-node",
		CpuPercent: 10.0,
		MemPercent: 50.0,
	}
	// Should not panic or alert below threshold
	alerts.check(metrics)
	if alerts.cpuHighSince != nil {
		t.Error("cpuHighSince should be nil when CPU is below threshold")
	}
}

func TestAlertStateTracksHighCPU(t *testing.T) {
	alerts := &AlertState{}
	metrics := Metrics{
		Hostname:   "test-node",
		CpuPercent: 90.0,
		MemPercent: 50.0,
	}
	alerts.check(metrics)
	if alerts.cpuHighSince == nil {
		t.Error("cpuHighSince should be set when CPU is above threshold")
	}
}

func TestAlertStateResetsWhenCPUDrops(t *testing.T) {
	alerts := &AlertState{}
	high := Metrics{CpuPercent: 90.0, MemPercent: 50.0}
	low := Metrics{CpuPercent: 10.0, MemPercent: 50.0}

	alerts.check(high)
	if alerts.cpuHighSince == nil {
		t.Error("cpuHighSince should be set after high CPU")
	}
	alerts.check(low)
	if alerts.cpuHighSince != nil {
		t.Error("cpuHighSince should be nil after CPU drops")
	}
}

func TestMetricsJSONParsing(t *testing.T) {
	raw := `{"timestamp":12345,"hostname":"test-node","cpu_percent":25.5,"mem_percent":60.0,"mem_used_mb":1024,"mem_total_mb":4096}`
	var metrics Metrics
	if err := json.Unmarshal([]byte(raw), &metrics); err != nil {
		t.Fatalf("Failed to parse metrics JSON: %v", err)
	}
	if metrics.Hostname != "test-node" {
		t.Errorf("Expected hostname 'test-node', got '%s'", metrics.Hostname)
	}
	if metrics.CpuPercent != 25.5 {
		t.Errorf("Expected CPU 25.5, got %f", metrics.CpuPercent)
	}
}

func TestAlertStateTracksHighMem(t *testing.T) {
	alerts := &AlertState{}
	metrics := Metrics{
		CpuPercent: 10.0,
		MemPercent: 95.0,
	}
	alerts.check(metrics)
	if alerts.memHighSince == nil {
		t.Error("memHighSince should be set when MEM is above threshold")
	}
}

func TestTimestampIsRecent(t *testing.T) {
	metrics := Metrics{Timestamp: uint64(time.Now().Unix())}
	now := uint64(time.Now().Unix())
	if metrics.Timestamp < now-5 || metrics.Timestamp > now+5 {
		t.Error("Timestamp should be within 5 seconds of now")
	}
}
