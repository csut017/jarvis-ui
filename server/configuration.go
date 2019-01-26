package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type monitorConfiguration struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

type appConfiguration struct {
	Sources    []monitorConfiguration `json:"sources"`
	DataPath   string                 `json:"dataPath"`
	StaticPath string                 `json:"staticPath"`
	Weather    *weatherConfiguration  `json:"weather"`
}

type weatherConfiguration struct {
	LocationCode     string `json:"location"`
	BaseURL          string `json:"url"`
	APIKey           string `json:"key"`
	RefreshPeriod    int64  `json:"refresh"`
	SunriseSunsetURL string `json:"sun"`
}

func readConfiguration(filePath string) (*appConfiguration, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("Unable to read configuration file:", err)
		return nil, err
	}

	var settings appConfiguration
	err = json.Unmarshal(file, &settings)
	if err != nil {
		log.Println("Unable to parse configuration file:", err)
		return nil, err
	}

	if settings.DataPath == "" {
		settings.DataPath = "data"
	}

	if settings.StaticPath == "" {
		settings.StaticPath = "static"
	}

	return &settings, nil
}
