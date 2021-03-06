package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type monitorConfiguration struct {
	Name       string `json:"name"`
	Port       string `json:"port"`
	IsDisabled bool   `json:"disabled"`
}

type roomConfiguration struct {
	Name       string   `json:"name"`
	Sources    []string `json:"sources"`
	Stations   []string `json:"stations"`
	IsDisabled bool     `json:"disabled"`
}

type stationConfiguration struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	IsDisabled bool   `json:"disabled"`
}

type weatherConfiguration struct {
	LocationCode     string `json:"location"`
	BaseURL          string `json:"url"`
	APIKey           string `json:"key"`
	RefreshPeriod    int64  `json:"refresh"`
	SunriseSunsetURL string `json:"sun"`
}

type appConfiguration struct {
	Rooms      []roomConfiguration    `json:"rooms"`
	Sources    []monitorConfiguration `json:"sources"`
	Stations   []stationConfiguration `json:"stations"`
	DataPath   string                 `json:"dataPath"`
	StaticPath string                 `json:"staticPath"`
	Weather    *weatherConfiguration  `json:"weather"`

	stations map[string]stationConfiguration
}

func (config *appConfiguration) FindStation(name string) *stationConfiguration {
	station, ok := config.stations[name]
	if !ok {
		return nil
	}
	return &station
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

	settings.stations = map[string]stationConfiguration{}
	for _, station := range settings.Stations {
		settings.stations[station.Name] = station
	}

	return &settings, nil
}
