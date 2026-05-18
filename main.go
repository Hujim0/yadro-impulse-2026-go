package main

import (
	"bufio"
	"dungeonGameLib/lib/engine"
	"dungeonGameLib/lib/parser"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type GameConfig struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

func loadConfigFromArgs() (*GameConfig, error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("usage: %s <config-file.json>", os.Args[0])
	}
	filePath := os.Args[1]

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var config GameConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	return &config, nil
}

func main() {
	config, err := loadConfigFromArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	gameInstance := engine.CreateGameInstance(config.Floors, config.Monsters, config.OpenAt, config.Duration)

	go gameInstance.RunGameLoop()

	go func() {
		for true {
			fmt.Println(<-gameInstance.OutputEventChan)
		}
	}()

	defer func() {
		for _, pair := range gameInstance.CompilePlayerData() {
			fmt.Println(pair.Player)
		}
	}()

	if err := processReader(os.Stdin, gameInstance.InputEventChan); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func processReader(r io.Reader, inputEventChan chan engine.GameEvent) error {
	reader := bufio.NewReader(r)
	lineNum := 0

	for {
		lineNum++
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if line == "" {
					break
				}
			} else {
				return fmt.Errorf("line %d: read error: %w", lineNum, err)
			}
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		event, err := parser.ParseLine(line)
		if err != nil {
			return fmt.Errorf("line %d: parse error: %w", lineNum, err)
		}

		inputEventChan <- event
	}

	return nil
}
