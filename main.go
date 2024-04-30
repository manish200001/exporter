package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	bandwidthGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "network_bandwidth_kbps", // Changed to "_kbps" for kilobits per second
		Help: "Measured network bandwidth in Kilobits per second (kbps)", // Updated help string
	})
	latencyGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "network_latency_ms", // No change for latency metric
		Help: "Measured network latency in milliseconds (ms)",
	})
	registry = prometheus.NewRegistry()
)

func init() {
	// Register the metrics with Prometheus
	registry.MustRegister(bandwidthGauge)
	registry.MustRegister(latencyGauge)
}

func main() {
	target := os.Getenv("TARGET_IP")
	if target == "" {
		fmt.Println("TARGET_IP environment variable is not set.")
		os.Exit(1)
	}

	// Start the HTTP server to expose metrics
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	go func() {
		err := http.ListenAndServe(":8330", nil)
		if err != nil {
			log.Fatal("Failed to start HTTP server:", err)
		}
	}()

	// Simulate gathering and updating metrics
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			// Simulate latency measurement
			latency, err := measureLatency(target)
			if err != nil {
				log.Println("Failed to measure latency:", err)
			} else {
				latencyGauge.Set(latency)
			}
			time.Sleep(5 * time.Second) // Update metrics every 5 seconds
		}
	}()

	go func() {
		defer wg.Done()
		for {
			// Simulate bandwidth measurement
			bandwidth, err := measureBandwidth(target)
			if err != nil {
				log.Println("Failed to measure bandwidth:", err)
			} else {
				bandwidthGauge.Set(bandwidth)
			}
			time.Sleep(5 * time.Second) // Update metrics every 5 seconds
		}
	}()

	wg.Wait() // Wait for goroutines to finish
}

func measureBandwidth(ipAddress string) (float64, error) {
	// Placeholder for bandwidth measurement logic
	cmd := exec.Command("iperf", "-c", ipAddress, "-t", "1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("iperf error: %v", err)
	}

	// Parse iperf output to get bandwidth
	lines := strings.Split(string(output), "\n")
	if len(lines) < 6 {
		return 0, fmt.Errorf("invalid iperf output")
	}
	bandwidthLine := lines[6]
	parts := strings.Fields(bandwidthLine)
	if len(parts) < 7 {
		return 0, fmt.Errorf("invalid iperf output")
	}
	bandwidthStr := parts[6]
	bandwidth, err := strconv.ParseFloat(bandwidthStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing error: %v", err)
	}
	return bandwidth, nil
}

func measureLatency(ipAddress string) (float64, error) {
	// Placeholder for latency measurement logic
	cmd := exec.Command("ping", "-c", "1", ipAddress)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ping error: %v", err)
	}

	// Extract latency from ping output
	latencyStr := string(output)
	latencyParts := strings.Split(latencyStr, "time=")
	if len(latencyParts) < 2 {
		return 0, fmt.Errorf("unable to parse latency")
	}
	latencyStr = strings.Split(latencyParts[1], " ")[0]
	latency, err := strconv.ParseFloat(latencyStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing error: %v", err)
	}
	return latency, nil
}

