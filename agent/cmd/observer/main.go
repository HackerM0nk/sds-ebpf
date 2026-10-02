package main

import (
	"context"
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
	"github.com/sds-ebpf/agent/pkg/events"
	"github.com/sds-ebpf/agent/pkg/exporter"
)

var (
	configFile   = flag.String("config", "", "Path to configuration file")
	outputDir    = flag.String("output", "/var/log/sds-observer", "Output directory for events")
	bpfObject    = flag.String("bpf-object", "", "Path to compiled eBPF object file")
	samplingOn   = flag.Int("sampling-on", 60, "Sampling duration in seconds")
	samplingOff  = flag.Int("sampling-off", 240, "Sampling off duration in seconds (5 min default)")
	httpEndpoint = flag.String("http", "", "Optional HTTP endpoint to send events")
	runOnce      = flag.Bool("run-once", false, "Run a single collection cycle and exit")
	verbose      = flag.Bool("verbose", false, "Verbose logging")
)

type eventCollector interface {
	Start(chan<- collector.Event) error
	Stop()
}

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

	builder, err := events.NewBuilder(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize canonical event builder: %v", err)
	}

	var httpExporter exporter.EventExporter
	if cfg.HTTPEndpoint != "" {
		httpExporter = exporter.NewHTTPExporter(cfg.HTTPEndpoint, nil)
	}

	log.Printf("Starting SDS eBPF Observer")
	log.Printf("Sampling: %ds ON, %ds OFF", cfg.SamplingOn, cfg.SamplingOff)
	log.Printf("Output: %s", cfg.OutputDir)

	if *runOnce {
		runCollection(col, cfg, builder, httpExporter)
		return
	}

	// Run sparse sampling loop
	ticker := time.NewTicker(time.Duration(cfg.SamplingOn+cfg.SamplingOff) * time.Second)
	defer ticker.Stop()

	// Start first collection immediately
	go runCollection(col, cfg, builder, httpExporter)

	for {
		select {
		case <-ticker.C:
			go runCollection(col, cfg, builder, httpExporter)
		case <-stop:
			log.Println("Shutting down...")
			return
		}
	}
}

func loadConfig() (*config.Config, error) {
	cfg := config.Default()
	cfg.OutputDir = *outputDir
	cfg.BPFObjectPath = *bpfObject
	cfg.SamplingOn = *samplingOn
	cfg.SamplingOff = *samplingOff
	cfg.HTTPEndpoint = *httpEndpoint
	cfg.Verbose = *verbose

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

func runCollection(col eventCollector, cfg *config.Config, builder *events.Builder, httpExporter exporter.EventExporter) {
	log.Printf("Starting collection cycle for %d seconds", cfg.SamplingOn)

	// Start collecting
	rawEvents := make(chan collector.Event, 1000)
	if err := col.Start(rawEvents); err != nil {
		log.Printf("Error starting collector: %v", err)
		return
	}

	// Create output file with timestamp
	timestamp := time.Now().Format("20060102_150405")
	outputFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("events_%s.ndjson", timestamp))

	file, err := os.Create(outputFile)
	if err != nil {
		log.Printf("Error creating output file: %v", err)
		col.Stop()
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	eventCount := 0

	// Collect for sampling duration
	timeout := time.After(time.Duration(cfg.SamplingOn) * time.Second)

	for {
		select {
		case event := <-rawEvents:
			canonicalEvent := builder.Build(event)
			if err := encoder.Encode(canonicalEvent); err != nil {
				log.Printf("Error encoding canonical event: %v", err)
			}
			eventCount++

			if cfg.Verbose {
				log.Printf("Event: %s [PID: %d]", canonicalEvent.Event.Kind, canonicalEvent.Process.PID)
			}

			if httpExporter != nil {
				go exportToHTTP(httpExporter, canonicalEvent)
			}

		case <-timeout:
			log.Printf("Collection cycle complete: %d events collected -> %s", eventCount, outputFile)
			col.Stop()
			return
		}
	}
}

func exportToHTTP(httpExporter exporter.EventExporter, event events.CanonicalEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpExporter.Export(ctx, event); err != nil {
		log.Printf("Error exporting event to HTTP: %v", err)
	}
}
