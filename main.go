package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// Init logger
	logger, err := initLogger()
	if err != nil {
		log.Fatalln("Failed to create logger: ", err)
	}
	defer logger.Sync()

	logger.Infoln("Starting application...")

	// Load config
	configPath := os.Getenv("APP_CONFIG")
	if configPath == "" {
		configPath = "config.json"
	}
	config, err := loadConfig(configPath)
	if err != nil {
		logger.Fatalln("Failed to create discord client: ", err)
	}

	// Start discord
	s, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		logger.Fatalf("Failed to create instance: ", err)
	}
	err = s.Open()
	if err != nil {
		logger.Fatalf("Failed to open session: ", err)
	}

	// Register commands
	commands, err := registerCommands(config, s, logger)
	if err != nil {
		logger.Fatalf("Failed to register commands: ", err)
	}
	defer unregisterCommands(config, s, commands)

	// Wait forever
	stopBot := make(chan os.Signal, 1)
	signal.Notify(stopBot, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stopBot

	logger.Infoln("Stopping application...")
}
