package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Metrics struct {
	Timestamp  uint64  `json:"timestamp"`
	CpuPercent float64 `json:"cpu_percent"`
	MemPercent float64 `json:"mem_percent"`
	MemUsedMB  uint64  `json:"mem_used_mb"`
	MemTotalMB uint64  `json:"mem_total_mb"`
}

type AlertState struct {
	cpuHighSince *time.Time
	memHighSince *time.Time
}

func (a *AlertState) check(metrics Metrics) {
	now := time.Now()

	// CPU alert: sustained above 20% for 10 seconds
	if metrics.CpuPercent > 80.0 {
		if a.cpuHighSince == nil {
			a.cpuHighSince = &now
		} else if time.Since(*a.cpuHighSince) >= 10*time.Second {
			fmt.Printf("[ALERT] CPU has been above 20%% for %s\n",
				time.Since(*a.cpuHighSince).Round(time.Second))
		}
	} else {
		a.cpuHighSince = nil
	}

	// MEM alert: sustained above 90% for 10 seconds
	if metrics.MemPercent > 90.0 {
		if a.memHighSince == nil {
			a.memHighSince = &now
		} else if time.Since(*a.memHighSince) >= 10*time.Second {
			fmt.Printf("[ALERT] MEM has been above 90%% for %s\n",
				time.Since(*a.memHighSince).Round(time.Second))
		}
	} else {
		a.memHighSince = nil
	}
}

func main() {
	socketPath := "/tmp/ironclad.sock"
	alerts := &AlertState{}

	for {
		fmt.Println("[ironclad-controller] Connecting to agent...")

		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			fmt.Printf("[ironclad-controller] Failed to connect: %v. Retrying in 2s...\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Println("[ironclad-controller] Connected to agent")
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			line := scanner.Text()
			var metrics Metrics
			if err := json.Unmarshal([]byte(line), &metrics); err != nil {
				fmt.Printf("[ironclad-controller] Failed to parse: %v\n", err)
				continue
			}

			fmt.Printf("[controller] ts=%d cpu=%.1f%% mem=%.1f%% (%dMB/%dMB)\n",
				metrics.Timestamp,
				metrics.CpuPercent,
				metrics.MemPercent,
				metrics.MemUsedMB,
				metrics.MemTotalMB,
			)

			alerts.check(metrics)
		}

		fmt.Println("[ironclad-controller] Agent disconnected, reconnecting...")
		conn.Close()
	}
}
