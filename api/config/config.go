package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	jsoniter "github.com/json-iterator/go"
)

// wanna use this globally
var (
	json = jsoniter.ConfigFastest
)

type Config struct {
	DBHost string `json:"db_host"`
	DBPort string `json:"db_port"`
	DBName string `json:"db_name"`
	DBUser string `json:"db_user"`
	DBPass string `json:"db_pass"`
}

// ParseConfigFile parses the API configuration file.
func ParseConfigFile(filepath string) (*Config, error) {
	config := &Config{}
	var json = jsoniter.ConfigFastest

	// read the config file
	// gonna bring in a filepath from the argument in ParseConfigFile and pass it into ReadFile
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not find confi.json fileat the given pathc %s", filepath)
	}

	// Try to unmarshal config file JSON into Config struct.
	if err := json.Unmarshal(file, config); err != nil {
		return nil, errors.New("Failed to Unmarshal JSON into struct")
	}

	return config, nil
}

