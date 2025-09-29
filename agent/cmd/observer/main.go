package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sds-ebpf/agent/pkg/collector"
	"github.com/sds-ebpf/agent/pkg/config"
)

var (
	configFile  = flag.String("config", "", "Path to configuration file")
	outputDir   = flag.String("output", "/var/log/sds-observer", "Output directory for events")
	samplingOn  = flag.Int("sampling-on", 60, "Sampling duration in seconds")
	samplingOff = flag.Int("sampling-off", 240, "Sampling off duration in seconds (5 min default)")
	httpEndpoint = flag.String("http", "", "Optional HTTP endpoint to send events")
	verbose     = flag.Bool("verbose", false, "Verbose logging")
)

func main() {
	flag.Parse()

	log.SetPrefix("[sds-observer] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Setup signal handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Initialize collector
	col, err := collector.NewCollector(cfg)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}
	defer col.Close()

	log.Printf("Starting SDS eBPF Observer")
	log.Printf("Sampling: %ds ON, %ds OFF", cfg.SamplingOn, cfg.SamplingOff)
	log.Printf("Output: %s", cfg.OutputDir)

	// Run sparse sampling loop
	ticker := time.NewTicker(time.Duration(cfg.SamplingOn+cfg.SamplingOff) * time.Second)
	defer ticker.Stop()

	// Start first collection immediately
	go runCollection(col, cfg)

	for {
		select {
		case <-ticker.C:
			go runCollection(col, cfg)
		case <-stop:
			log.Println("Shutting down...")
			return
		}
	}
}

func loadConfig() (*config.Config, error) {
	cfg := &config.Config{
		OutputDir:    *outputDir,
		SamplingOn:   *samplingOn,
		SamplingOff:  *samplingOff,
		HTTPEndpoint: *httpEndpoint,
		Verbose:      *verbose,
	}

	if *configFile != "" {
		file, err := os.ReadFile(*configFile)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err := json.Unmarshal(file, cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	return cfg, nil
}

func runCollection(col *collector.Collector, cfg *config.Config) {
	log.Printf("Starting collection cycle for %d seconds", cfg.SamplingOn)

	// Start collecting
	events := make(chan collector.Event, 1000)
	go col.Start(events)

	// Create output file with timestamp
	timestamp := time.Now().Format("20060102_150405")
	outputFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("events_%s.json", timestamp))

	file, err := os.Create(outputFile)
	if err != nil {
		log.Printf("Error creating output file: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	eventCount := 0

	// Collect for sampling duration
	timeout := time.After(time.Duration(cfg.SamplingOn) * time.Second)

	for {
		select {
		case event := <-events:
			if err := encoder.Encode(event); err != nil {
				log.Printf("Error encoding event: %v", err)
			}
			eventCount++

			if cfg.Verbose {
				log.Printf("Event: %s [PID: %d]", event.Type, event.PID)
			}

			// Optionally send to HTTP endpoint
			if cfg.HTTPEndpoint != "" {
				go sendToHTTP(cfg.HTTPEndpoint, event)
			}

		case <-timeout:
			log.Printf("Collection cycle complete: %d events collected -> %s", eventCount, outputFile)
			col.Stop()
			return
		}
	}
}

func sendToHTTP(endpoint string, event collector.Event) {
	// Implement HTTP posting logic here
	// This is a placeholder for Phase 2 integration
	_ = endpoint
	_ = event
}