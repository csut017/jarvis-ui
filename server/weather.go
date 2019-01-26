package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type weatherService struct {
	current       *CurrentWeather
	forecast      *WeatherForecast
	sunriseSunset *SunriseSunset
	downloaded    time.Time
	mutex         sync.Mutex
	isRunning     bool
	stopRequest   chan int
	stopReply     chan int
}

func (service *weatherService) Start(config *weatherConfiguration) error {
	if service.isRunning {
		return nil
	}

	log.Printf("[Weather] Starting service")
	service.stopRequest = make(chan int)
	service.stopReply = make(chan int)

	go service.process(true, config)
	service.isRunning = true
	return nil
}

func (service *weatherService) Stop(timeOut time.Duration) error {
	if !service.isRunning {
		return nil
	}

	log.Printf("[Weather] Stopping service")
	service.isRunning = false
	service.stopRequest <- 1
	select {
	case <-service.stopReply:
		return nil
	case <-time.After(timeOut):
		return errors.New("Stop weather service timed out")
	}
}

func (service *weatherService) process(isFirst bool, config *weatherConfiguration) {

	if isFirst {
		service.Download(config)
	}

	stop := false
	select {
	case <-service.stopRequest:
		service.stopReply <- 1
		stop = true

	case <-time.After(time.Duration(int64(time.Minute) * config.RefreshPeriod)):
		err := service.Download(config)
		if err != nil {
			log.Printf("[Weather] Unable to update weather information: %v", err)
		} else {
			log.Printf("[Weather] Weather information updated")
		}
	}
	if !stop {
		service.process(false, config)
	}
}

func (service *weatherService) GetCurrentWeather() *CurrentWeather {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	weather := service.current
	clone := *weather
	return &clone
}

func (service *weatherService) GetWeatherForecast() *WeatherForecast {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	forecast := service.forecast
	clone := *forecast
	return &clone
}

func (service *weatherService) GetSunriseSunset() *SunriseSunset {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	light := service.sunriseSunset
	clone := *light
	return &clone
}

func (service *weatherService) Download(config *weatherConfiguration) error {
	log.Printf("[Weather] Downloading weather information")
	weather, err := service.downloadCurrent(config.LocationCode, config)
	if err != nil {
		return err
	}

	forecast, err := service.downloadForecast(config.LocationCode, config)
	if err != nil {
		return err
	}

	log.Printf("[Weather] Downloading sunrise and sunset information")
	sunriseSunset, err := service.downloadSunriseSunset(weather.Coordinates, config)
	if err != nil {
		return err
	}

	log.Printf("[Weather] Storing weather, sunrise and sunset")
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.current = weather
	service.forecast = forecast
	service.sunriseSunset = sunriseSunset
	service.downloaded = time.Now()

	return nil
}

func (service *weatherService) downloadCurrent(cityCode string, config *weatherConfiguration) (*CurrentWeather, error) {
	log.Printf("[Weather] Downloading current weather")
	client := http.Client{}
	resp, err := client.Get(config.BaseURL + "weather?id=" + cityCode + "&units=metric&APPID=" + config.APIKey)
	if err != nil {
		log.Printf("[Weather] Unable to download current weather: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var weather CurrentWeather
	err = decoder.Decode(&weather)
	if err != nil {
		log.Printf("[Weather] Unable to parse current weather: %v", err)
		return nil, err
	}

	return &weather, nil
}

func (service *weatherService) downloadForecast(cityCode string, config *weatherConfiguration) (*WeatherForecast, error) {
	log.Printf("[Weather] Downloading weather forecast")
	client := http.Client{}
	resp, err := client.Get(config.BaseURL + "forecast?id=" + cityCode + "&units=metric&APPID=" + config.APIKey)
	if err != nil {
		log.Printf("[Weather] Unable to download weather forecast: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var forecast WeatherForecast
	err = decoder.Decode(&forecast)
	if err != nil {
		log.Printf("[Weather] Unable to parse weather forecast: %v", err)
		return nil, err
	}

	return &forecast, nil
}

func (service *weatherService) downloadSunriseSunset(coordinates WeatherCoordinates, config *weatherConfiguration) (*SunriseSunset, error) {
	log.Printf("[Weather] Downloading sunrise and sunset")
	client := http.Client{}
	url := fmt.Sprintf(config.SunriseSunsetURL+"?lat=%f&lng=%f&formatted=0", coordinates.Latitude, coordinates.Longitude)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("[Weather] Unable to download sunrise and sunset: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var results struct {
		Results SunriseSunset `json:"results"`
		Status  string        `json:"status"`
	}
	err = decoder.Decode(&results)
	if err != nil {
		log.Printf("[Weather] Unable to parse sunrise and sunset: %v", err)
		return nil, err
	}

	if results.Status != "OK" {
		log.Printf("[Weather] Invalid sunrise and sunset: status=%s", results.Status)
		return nil, err
	}

	return &results.Results, nil
}

type WeatherCoordinates struct {
	Longitude float32 `json:"lon"`
	Latitude  float32 `json:"lat"`
}

type CurrentWeather struct {
	Coordinates WeatherCoordinates   `json:"coord"`
	Weather     []WeatherInformation `json:"weather"`
	Main        MainInformation      `json:"main"`
	Wind        WindInformation      `json:"wind"`
	System      struct {
		Sunrise int64 `json:"sunrise"`
		Sunset  int64 `json:"sunset"`
	} `json:"sys"`
}

type WeatherForecast struct {
	List []ThreeHourForecast `json:"list"`
}

type ThreeHourForecast struct {
	Weather     []WeatherInformation `json:"weather"`
	Main        MainInformation      `json:"main"`
	Wind        WindInformation      `json:"wind"`
	DateAndTime string               `json:"dt_txt"`
}

type WeatherInformation struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

type MainInformation struct {
	Temperature        float64 `json:"temp"`
	Pressure           float64 `json:"pressure"`
	Humidity           float64 `json:"humidity"`
	MinimumTemperature float64 `json:"temp_min"`
	MaximumTemperature float64 `json:"temp_max"`
}

type WindInformation struct {
	Speed     float64 `json:"speed"`
	Direction float64 `json:"deg"`
}

type SunriseSunset struct {
	AstronomicalTwilightBegin string `json:"astronomical_twilight_begin"`
	CivilTwilightBegin        string `json:"civil_twilight_begin"`
	Sunrise                   string `json:"sunrise"`
	Sunset                    string `json:"sunset"`
	CivilTwilightEnd          string `json:"civil_twilight_end"`
	AstronomicalTwilightEnd   string `json:"astronomical_twilight_end"`
}
