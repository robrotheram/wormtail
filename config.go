package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Tailscale struct {
		APIKey   string `yaml:"api_key"`
		Hostnmae string `yaml:"hostnmae"`
	} `yaml:"tailscale"`
	Routes []Route `yaml:"routes"`
}

func LoadConfig() Config {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Unmarshal the YAML data into the Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return config
}
