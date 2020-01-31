package config

import (
	"encoding/json"
	"os"
)

//Configuration struct that is used to retrieve configuration values
type Configuration struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// ReadConfiguration retrieves the configration values from config file
func ReadConfiguration() (Configuration, error) {
	var configuration Configuration

	file, err := os.Open("./infra/config/truelayer/config.json")
	if err != nil {
		return configuration, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&configuration)
	if err != nil {
		return configuration, err
	}

	return configuration, nil
}
