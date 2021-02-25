package main

import (
	"encoding/json"
	"io/ioutil"
)

var config struct {
	DatabaseUser string `json:"databaseUser"`
	DatabasePass string `json:"databasePass"`
	DatabaseHost string `json:"databaseHost"`
	DatabasePort string `json:"databasePort"`
	DatabaseName string `json:"databaseName"`
	Port         string `json:"port"`
}

func readConfig() error {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	if config.DatabaseHost == "" {
		config.DatabaseHost = "localhost"
	}

	if config.DatabasePort == "" {
		config.DatabasePort = "3306"
	}

	if config.Port == "" {
		config.Port = "80"
	}

	return nil
}
