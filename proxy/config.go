// Copyright 2017 Landonia Ltd. All rights reserved.

package proxy

import (
	"bytes"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Configuration wraps the settings required for the app
type Configuration struct {
	Addr      string       `yaml:"addr"`     // The host to locally bind
	LogLevel  string       `yaml:"loglevel"` // The log level to use
	StaticDir string       `yaml:"static"`   // The static hosts root directory
	Proxies   []HostConfig `yaml:"proxies"`  // The proxy information
}

// HostConfig information
type HostConfig struct {
	ProxyPath string `yaml:"proxy"`
	To        string `yaml:"to"`
}

// DefaultConfig will return a sensible default configuration
func DefaultConfig() Configuration {
	conf := Configuration{}
	conf.Addr = DefaultServerAddr
	conf.StaticDir = "."
	conf.LogLevel = "DEBUG"
	return conf
}

// ParseFileConfig will return a new Configuration
func ParseFileConfig(path string) (Configuration, error) {

	// try opening the file to see if it exists
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return Configuration{}, err
	}
	conf := Configuration{}
	var b bytes.Buffer
	_, err = b.ReadFrom(file)
	if err == nil {
		err = yaml.Unmarshal(b.Bytes(), &conf)
	}
	return conf, err
}
