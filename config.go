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
	Metadata struct {
		Name    string `yaml:"name"`
		LogoURL string `yaml:"logoURL"`
	}
	Services []Service
}

type Service struct {
	Name  string `yaml:"name"`
	Index string `yaml:"index"`
}

func (config *Config) loadConfig(path string) *Config {
	configData, err := ioutil.ReadFile(path)

	err = yaml.Unmarshal(configData, &config)

	if err != nil {
		fmt.Printf("error: %+v", err)
	}

	return config
}
