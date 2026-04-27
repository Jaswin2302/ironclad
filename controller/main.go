package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Metrics struct {
	Timestamp   uint64  `json:"timestamp"`
	CpuPercent  float64 `json:"cpu_percent"`
	MemPercent  float64 `json:"mem_percent"`
	MemUsedMB   uint64  `json:"mem_used_mb"`
	MemTotalMB  uint64  `json:"mem_total_mb"`
}

func main() {
	socketPath := "/tmp/ironclad.sock"

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
		}

		fmt.Println("[ironclad-controller] Agent disconnected, reconnecting...")
		conn.Close()
	}
}