package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

type GameConfig struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

func LoadConfigFromFile(filePath string) (*GameConfig, error) {
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
