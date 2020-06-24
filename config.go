package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Elasticsearch struct {
		Hostname string `yaml:"hostname"`
		Port     string `yaml:"port"`
	}
	Server struct {
		Hostname string `yaml:"hostname"`
		Port     string `yaml:"port"`
	}
	Services []Service
}

func (config *Config) loadConfig() *Config {
	configData, err := ioutil.ReadFile("config.yml")

	err = yaml.Unmarshal(configData, &config)

	if err != nil {
		fmt.Printf("error: %v", err)
	}

	return config
}
