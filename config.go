package main

import (
	"encoding/json"
	"os"
)

type AppConfig struct {
	Token    string
	GuildID  string
	StartCmd string
	StopCmd  string
	CheckCmd string
}

func loadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &AppConfig{}
	err = json.Unmarshal(data, config)
	return config, err
}
