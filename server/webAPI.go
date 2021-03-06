package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type webAPI struct {
	Router   *mux.Router
	addr     string
	data     *dataStore
	monitors *monitorStore
	config   *appConfiguration
	upgrader websocket.Upgrader
	hub      *websocketHub
	weather  *weatherService
}

type itemStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type sourceDetails struct {
	Name      string   `json:"name,omitempty"`
	Sensors   []string `json:"sensors"`
	Effectors []string `json:"effectors"`
}

func newWebAPI(addr string, data *dataStore, monitors *monitorStore, weather *weatherService, config *appConfiguration) (*webAPI, error) {
	api := webAPI{
		addr:     addr,
		data:     data,
		monitors: monitors,
		weather:  weather,
		config:   config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		hub: newHub(),
	}
	api.Router = api.initialise(addr)
	return &api, nil
}

func (api *webAPI) initialise(addr string) *mux.Router {
	router := mux.NewRouter().PathPrefix("/api").Subrouter()
	router.KeepContext = true

	// Methods for working with sources
	router.HandleFunc("/sources", api.listSources).Methods("GET")
	router.HandleFunc("/sources/{source}", api.getSourceDetails).Methods("GET")
	router.HandleFunc("/sources/{source}/values", api.listSourceValues).Methods("GET")
	router.HandleFunc("/sources/{source}/sensors", api.listSourceOutput).Methods("GET")
	router.HandleFunc("/sources/{source}/effectors", api.listSourceInput).Methods("GET")
	router.HandleFunc("/sources/{source}/effectors", api.processSourceCommand).Methods("POST")

	// Methods for generating speech
	router.HandleFunc("/speech", api.generateSpeechFromGET).Methods("GET")
	router.HandleFunc("/speech", api.generateSpeechFromPOST).Methods("POST")

	// Methods for working with stations
	router.HandleFunc("/stations", api.getStations).Methods("GET")
	router.HandleFunc("/stations/{station}", api.getStationDetails).Methods("GET")
	router.HandleFunc("/stations/{station}/sources/{source}/values", api.listStationSourceValues).Methods("GET")
	router.HandleFunc("/stations/{station}/sources/{source}/effectors", api.processStationSourceCommand).Methods("POST")

	// Methods for retrieving room information
	router.HandleFunc("/rooms", api.getRooms).Methods("GET")
	router.HandleFunc("/rooms/{room}", api.getRoomDetails).Methods("GET")

	// Methods for retrieving weather information
	router.HandleFunc("/weather", api.getWeather).Methods("GET")
	router.HandleFunc("/weather/raw", api.getRawWeather).Methods("GET")
	router.HandleFunc("/sun", api.getSunriseSunset).Methods("GET")

	// Methods for initialising a websocket
	router.HandleFunc("/ws", api.startWebsocket).Methods("GET")

	return router
}

func (api *webAPI) start() {
	go api.hub.run()
}

func (api *webAPI) writeErrorJSON(resp http.ResponseWriter, statusCode int, message string) {
	api.writeStatusJSON(resp, statusCode, "error", message)
}

func (api *webAPI) writeStatusJSON(resp http.ResponseWriter, statusCode int, status string, message string) {
	resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
	msg := struct {
		Status  string `json:"status"`
		Message string `json:"msg"`
	}{
		Status:  status,
		Message: message,
	}
	resp.WriteHeader(statusCode)
	if err := json.NewEncoder(resp).Encode(msg); err != nil {
		log.Printf("[API] Encoding error: %v", err)
		http.Error(resp, http.StatusText(500), 500)
	}
}

func (api *webAPI) writeDataJSON(resp http.ResponseWriter, statusCode int, data interface{}) {
	resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp.WriteHeader(statusCode)
	if err := json.NewEncoder(resp).Encode(data); err != nil {
		log.Printf("[API] Encoding error: %v", err)
		http.Error(resp, http.StatusText(500), 500)
	}
}

func (api *webAPI) listSources(resp http.ResponseWriter, req *http.Request) {
	log.Printf("[API] Listing sources")
	out := struct {
		Items []itemStatus `json:"sources"`
	}{
		Items: make([]itemStatus, len(api.config.Sources)),
	}

	for pos, source := range api.config.Sources {
		status := "Disabled"
		if !source.IsDisabled {
			status = "Active"
		}
		out.Items[pos] = itemStatus{
			Name:   source.Name,
			Status: status,
		}
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) retrieveSource(resp http.ResponseWriter, req *http.Request) (string, *monitor) {
	vars := mux.Vars(req)
	name := vars["source"]
	store := api.monitors.Get(name)
	if store == nil {
		log.Printf("[API] Cannot find source %s", name)
		api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Unknown source")
	}
	return name, store
}

func (api *webAPI) listSourceValues(resp http.ResponseWriter, req *http.Request) {
	name, store := api.retrieveSource(resp, req)
	if store == nil {
		return
	}

	countText := "100"
	counts, ok := req.URL.Query()["count"]
	if ok {
		countText = counts[0]
	}

	count, err := strconv.Atoi(countText)
	if err != nil {
		count = 100
	}

	log.Printf("[API] Listing data for source %s", name)
	items := api.data.GetLast(name, count)
	out := struct {
		Count int              `json:"count"`
		Items *[]monitorResult `json:"items"`
	}{
		Count: len(*items),
		Items: items,
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) getSourceDetails(resp http.ResponseWriter, req *http.Request) {
	name, store := api.retrieveSource(resp, req)
	if store == nil {
		return
	}

	log.Printf("[API] Getting details for source %s", name)
	out := sourceDetails{
		Sensors:   store.InputTypes(),
		Effectors: store.OutputTypes(),
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) listSourceOutput(resp http.ResponseWriter, req *http.Request) {
	name, store := api.retrieveSource(resp, req)
	if store == nil {
		return
	}

	log.Printf("[API] Listing sensors for source %s", name)
	out := struct {
		Items []string `json:"items"`
	}{
		Items: store.InputTypes(),
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) listSourceInput(resp http.ResponseWriter, req *http.Request) {
	name, store := api.retrieveSource(resp, req)
	if store == nil {
		return
	}

	log.Printf("[API] Listing effectors for source %s", name)
	out := struct {
		Items []string `json:"items"`
	}{
		Items: store.OutputTypes(),
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) processSourceCommand(resp http.ResponseWriter, req *http.Request) {
	name, store := api.retrieveSource(resp, req)
	if store == nil {
		return
	}

	cmd := &command{}
	err := json.NewDecoder(req.Body).Decode(cmd)
	if err != nil {
		log.Printf("[API] ERROR: Unable to parse incoming JSON: %v", err)
		api.writeStatusJSON(resp, http.StatusBadRequest, "Error", "Invalid command")
		return
	}

	if err = store.SendCommand(cmd); err != nil {
		msg := fmt.Sprintf("Unable to send command: %v", err)
		log.Printf("[API] ERROR: " + msg)
		api.writeStatusJSON(resp, http.StatusBadRequest, "Failure", msg)
		return
	}

	log.Printf("[API] Sending %s action to %s in source %s", cmd.Action, cmd.Name, name)
	api.writeStatusJSON(resp, http.StatusOK, "Ok", "Command sent")
}

func (api *webAPI) generateSpeechFromGET(resp http.ResponseWriter, req *http.Request) {
	args := req.URL.Query()
	text := strings.Join(args["text"], " ")
	voice := strings.Join(args["voice"], " ")
	format := strings.Join(args["format"], " ")
	api.generateSpeech(resp, req, text, voice, format)
}

func (api *webAPI) generateSpeechFromPOST(resp http.ResponseWriter, req *http.Request) {
	cmd := &struct {
		Text   string `json:"text"`
		Voice  string `json:"voice"`
		Format string `json:"format"`
	}{}
	err := json.NewDecoder(req.Body).Decode(cmd)
	if err != nil {
		log.Printf("[API] ERROR: Unable to parse incoming JSON: %v", err)
		api.writeStatusJSON(resp, http.StatusBadRequest, "Error", "Invalid command")
		return
	}
	api.generateSpeech(resp, req, cmd.Text, cmd.Voice, cmd.Format)
}

func (api *webAPI) generateSpeech(resp http.ResponseWriter, req *http.Request, text, voice, format string) {
	if text == "" {
		log.Printf("[API] ERROR: No text to speak")
		api.writeStatusJSON(resp, http.StatusBadRequest, "Error", "Missing text")
		return
	}

	if voice == "" {
		voice = "neutral"
	} else {
		voice = strings.ToLower(voice)
	}
	if format == "" {
		format = "mp3"
	} else {
		format = strings.ToLower(format)
	}
	log.Printf("[API] Saying speech '%s' with %s voice to %s", text, voice, format)
	audio, err := generateSpeech(text, voice, format)
	if err != nil {
		log.Printf("[API] ERROR: Unable to generate speech: %v", err)
		api.writeStatusJSON(resp, http.StatusBadRequest, "Failure", "Unable to generate speech")
		return
	}

	resp.Header().Set("Content-Disposition", "attachment; filename=speech."+format)
	resp.Header().Set("Content-Type", "audio/mpeg")
	resp.Header().Set("Content-Length", strconv.Itoa(len(audio.AudioContent)))
	resp.Write(audio.AudioContent)
}

func (api *webAPI) getStations(resp http.ResponseWriter, req *http.Request) {
	log.Printf("[API] Listing stations")
	local := 0
	if len(api.config.Sources) > 0 {
		local = 1
	}
	out := struct {
		Items []itemStatus `json:"items"`
	}{
		Items: make([]itemStatus, len(api.config.Stations)+local),
	}
	if local == 1 {
		out.Items[0] = itemStatus{
			Name:   "local",
			Status: "Active",
		}
	}
	for pos, station := range api.config.Stations {
		status := "Disabled"
		if !station.IsDisabled {
			status = "Active"
		}
		out.Items[pos+local] = itemStatus{
			Name:   station.Name,
			Status: status,
		}
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) getStationDetails(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["station"]
	if name == "local" {
		log.Printf("[API] Generating local station information")
		out := struct {
			Sources []sourceDetails `json:"sources"`
		}{
			Sources: make([]sourceDetails, len(*api.monitors)),
		}
		pos := 0
		for name, store := range *api.monitors {
			out.Sources[pos] = sourceDetails{
				Name:      name,
				Sensors:   store.InputTypes(),
				Effectors: store.OutputTypes(),
			}
			pos++
		}

		api.writeDataJSON(resp, http.StatusOK, out)
	} else {
		station := api.config.FindStation(name)

		if station == nil {
			log.Printf("[API] Cannot find station %s", name)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Unknown station")
			return
		}

		res, err := http.Get("http://" + station.Address + "/api/stations/local")
		if err != nil {
			log.Printf("[API] Cannot query station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		} else if res.StatusCode != http.StatusOK {
			err = fmt.Errorf("Unable to retrieve from station: %s", res.Status)
			log.Printf("[API] Cannot query station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		}

		log.Printf("[API] Generating station information for %s", name)
		out := struct {
			Station string          `json:"station"`
			Sources []sourceDetails `json:"sources"`
		}{}
		if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
			log.Printf("[API] Cannot decode JSON from station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		}
		out.Station = name
		api.writeDataJSON(resp, http.StatusOK, out)
	}
}

func (api *webAPI) listStationSourceValues(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["station"]
	if name == "local" {
		log.Printf("[API] Generating local station values")
		api.listSourceValues(resp, req)
	} else {
		station := api.config.FindStation(name)
		if station == nil {
			log.Printf("[API] Cannot find station %s", name)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Unknown station")
			return
		}

		sourceName := vars["source"]
		res, err := http.Get("http://" + station.Address + "/api/sources/" + sourceName + "/values")
		if err != nil {
			log.Printf("[API] Cannot query station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		} else if res.StatusCode != http.StatusOK {
			err = fmt.Errorf("Unable to retrieve from station: %s", res.Status)
			log.Printf("[API] Cannot query station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		}

		log.Printf("[API] Generating station values from %s", name)
		out := struct {
			Station string           `json:"station"`
			Count   int              `json:"count"`
			Items   *[]monitorResult `json:"items"`
		}{}
		if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
			log.Printf("[API] Cannot decode JSON from station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		}
		out.Station = name
		api.writeDataJSON(resp, http.StatusOK, out)
	}
}

func (api *webAPI) processStationSourceCommand(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["station"]
	if name == "local" {
		log.Printf("[API] Processing local station command")
		api.processSourceCommand(resp, req)
	} else {
		station := api.config.FindStation(name)
		if station == nil {
			log.Printf("[API] Cannot find station %s", name)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Unknown station")
			return
		}

		sourceName := vars["source"]
		res, err := http.Post(
			"http://"+station.Address+"/api/sources/"+sourceName+"/effectors",
			"application/json",
			req.Body)
		if err != nil {
			log.Printf("[API] Cannot post command to station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		} else if res.StatusCode != http.StatusOK {
			err = fmt.Errorf("Unable to post command to station: %s", res.Status)
			log.Printf("[API] Cannot query station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		}

		log.Printf("[API] Retrieving command result from %s", name)
		out := struct {
			Station string `json:"station"`
			Status  string `json:"status"`
			Message string `json:"msg"`
		}{}
		if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
			log.Printf("[API] Cannot decode JSON from station %s: %v", name, err)
			api.writeStatusJSON(resp, http.StatusNotFound, "Error", "Station not available")
			return
		}
		out.Station = name
		api.writeDataJSON(resp, http.StatusOK, out)
	}
}

func (api *webAPI) getRooms(resp http.ResponseWriter, req *http.Request) {
	log.Printf("[API] Listing rooms")
	out := struct {
		Items []itemStatus `json:"items"`
	}{
		Items: make([]itemStatus, len(api.config.Rooms)),
	}
	for pos, room := range api.config.Rooms {
		status := "Disabled"
		if !room.IsDisabled {
			status = "Active"
		}
		out.Items[pos] = itemStatus{
			Name:   room.Name,
			Status: status,
		}
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) getRoomDetails(resp http.ResponseWriter, req *http.Request) {
	name, store := api.retrieveSource(resp, req)
	if store == nil {
		return
	}

	log.Printf("[API] Generating room information for %s", name)
	out := struct {
		Summary string `json:"summary"`
	}{
		Summary: "TODO",
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) getWeather(resp http.ResponseWriter, req *http.Request) {
	oneWord := ""
	current := api.weather.GetCurrentWeather()
	var currentTemp, minTemp, maxTemp float64

	currentWeather := "There is no weather information"
	if current != nil && len(current.Weather) > 0 {
		oneWord = current.Weather[0].Main
		currentWeather = fmt.Sprintf(
			"The current weather is %s, the temperature is %.f°C",
			current.Weather[0].Description,
			current.Main.Temperature)
		currentTemp = current.Main.Temperature
	}

	forecast := api.weather.GetWeatherForecast()
	if forecast == nil {
		api.writeStatusJSON(resp, http.StatusServiceUnavailable, "Not available", "Weather information has not been downloaded")
		return
	}

	weatherForecast, listLen := "There is no forecast", len(forecast.List)
	if forecast != nil && listLen > 0 {
		item := forecast.List[0]
		minTemp, maxTemp = item.Main.MinimumTemperature, item.Main.MaximumTemperature
		for loop := 1; loop < 8 && loop < listLen; loop++ {
			temp := forecast.List[loop].Main
			if minTemp > temp.MinimumTemperature {
				minTemp = temp.MinimumTemperature
			}
			if maxTemp < temp.MaximumTemperature {
				maxTemp = temp.MaximumTemperature
			}
		}
		if math.Abs(minTemp-maxTemp) > 0.5 {
			weatherForecast = fmt.Sprintf(
				"The forecast is %s, with a temperature between %.f and %.f°C",
				item.Weather[0].Description,
				minTemp,
				maxTemp)
		} else {
			weatherForecast = fmt.Sprintf(
				"The forecast is %s, with a temperature of %.f°C",
				item.Weather[0].Description,
				maxTemp)
		}
	}

	item := struct {
		Current     string `json:"current"`
		OneWord     string `json:"oneWord"`
		Forecast    string `json:"forecast"`
		Temperature struct {
			Minimum float64 `json:"min"`
			Current float64 `json:"current"`
			Maximum float64 `json:"max"`
		} `json:"temperature"`
	}{
		Current:  currentWeather,
		OneWord:  oneWord,
		Forecast: weatherForecast,
	}
	item.Temperature.Minimum = minTemp
	item.Temperature.Current = currentTemp
	item.Temperature.Maximum = maxTemp
	api.writeDataJSON(resp, 200, item)
}

func (api *webAPI) getRawWeather(resp http.ResponseWriter, req *http.Request) {
	item := struct {
		Current  interface{} `json:"current"`
		Forecast interface{} `json:"forecast"`
	}{
		Current:  api.weather.GetCurrentWeather(),
		Forecast: api.weather.GetWeatherForecast(),
	}
	api.writeDataJSON(resp, 200, item)
}

func (api *webAPI) getSunriseSunset(resp http.ResponseWriter, req *http.Request) {
	item := struct {
		Results interface{} `json:"results"`
	}{
		Results: api.weather.GetSunriseSunset(),
	}
	api.writeDataJSON(resp, 200, item)
}

func (api *webAPI) startWebsocket(resp http.ResponseWriter, req *http.Request) {
	log.Printf("[API] Starting websocket connection")
	conn, err := api.upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Printf("[API] Unable to upgrade websocket connection: %v", err)
		return
	}

	client := &websocketClient{hub: api.hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
}
