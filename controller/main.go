package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Timestamp  uint64  `json:"timestamp"`
	Hostname   string  `json:"hostname"`
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
	if metrics.CpuPercent > 80.0 {
		if a.cpuHighSince == nil {
			a.cpuHighSince = &now
		} else if time.Since(*a.cpuHighSince) >= 10*time.Second {
			fmt.Printf("[ALERT] [%s] CPU has been above 80%% for %s\n",
				metrics.Hostname,
				time.Since(*a.cpuHighSince).Round(time.Second))
		}
	} else {
		a.cpuHighSince = nil
	}
	if metrics.MemPercent > 90.0 {
		if a.memHighSince == nil {
			a.memHighSince = &now
		} else if time.Since(*a.memHighSince) >= 10*time.Second {
			fmt.Printf("[ALERT] [%s] MEM has been above 90%% for %s\n",
				metrics.Hostname,
				time.Since(*a.memHighSince).Round(time.Second))
		}
	} else {
		a.memHighSince = nil
	}
}

func main() {
	socketPath := "/tmp/ironclad.sock"
	alerts := &AlertState{}

	cpuGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ironclad_cpu_percent",
		Help: "Current CPU usage percentage",
	}, []string{"host"})

	memGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ironclad_mem_percent",
		Help: "Current memory usage percentage",
	}, []string{"host"})

	memUsedGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ironclad_mem_used_mb",
		Help: "Current memory used in MB",
	}, []string{"host"})

	prometheus.MustRegister(cpuGauge, memGauge, memUsedGauge)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println("[ironclad-controller] Prometheus metrics at :9100/metrics")
		http.ListenAndServe(":9100", nil)
	}()

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
			fmt.Printf("[controller] [%s] ts=%d cpu=%.1f%% mem=%.1f%% (%dMB/%dMB)\n",
				metrics.Hostname,
				metrics.Timestamp,
				metrics.CpuPercent,
				metrics.MemPercent,
				metrics.MemUsedMB,
				metrics.MemTotalMB,
			)

			cpuGauge.WithLabelValues(metrics.Hostname).Set(metrics.CpuPercent)
			memGauge.WithLabelValues(metrics.Hostname).Set(metrics.MemPercent)
			memUsedGauge.WithLabelValues(metrics.Hostname).Set(float64(metrics.MemUsedMB))

			alerts.check(metrics)
		}
		fmt.Println("[ironclad-controller] Agent disconnected, reconnecting...")
		conn.Close()
	}
}
