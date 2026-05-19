package main

import (
	"bufio"
	"dungeonGameLib/internal/monitor"
	"dungeonGameLib/lib/engine/event"
	"dungeonGameLib/lib/parser"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <config.json> < events\n", os.Args[0])
		os.Exit(1)
	}

	cfgData, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read config: %v\n", err)
		os.Exit(1)
	}
	var cfg monitor.Config
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "parse config: %v\n", err)
		os.Exit(1)
	}

	evts, err := readEvents(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read events: %v\n", err)
		os.Exit(1)
	}

	monitor.Run(cfg, evts)
}

func readEvents(r io.Reader) ([]event.GameEvent, error) {
	var evts []event.GameEvent
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		ev, err := parser.ParseLine(line)
		if err != nil {
			return nil, fmt.Errorf("parse: %w", err)
		}
		evts = append(evts, ev)
	}
	return evts, scanner.Err()
}