package main

import (
	"encoding/json"
	"log"
	"net/http"

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
}

func newWebAPI(addr string, data *dataStore, monitors *monitorStore, config *appConfiguration) (*webAPI, error) {
	api := webAPI{
		addr:     addr,
		data:     data,
		monitors: monitors,
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
	router.HandleFunc("/sources/{name}/values", api.listSourceValues).Methods("GET")
	router.HandleFunc("/sources/{name}/sensors", api.listSourceOutput).Methods("GET")
	router.HandleFunc("/sources/{name}/effectors", api.listSourceInput).Methods("GET")
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
		Sources []monitorConfiguration `json:"sources"`
	}{
		Sources: api.config.Sources,
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) listSourceValues(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	store := api.monitors.Get(name)
	if store == nil {
		log.Printf("[API] Cannot find source %s", name)
		api.writeDataJSON(resp, http.StatusNotFound, "Unknown source")
		return
	}

	log.Printf("[API] Listing data for source %s", name)
	out := struct {
		Items *[]monitorResult `json:"items"`
	}{
		Items: api.data.GetItems(name),
	}
	api.writeDataJSON(resp, http.StatusOK, out)
}

func (api *webAPI) listSourceOutput(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	store := api.monitors.Get(name)
	if store == nil {
		log.Printf("[API] Cannot find source %s", name)
		api.writeDataJSON(resp, http.StatusNotFound, "Unknown source")
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
	vars := mux.Vars(req)
	name := vars["name"]
	store := api.monitors.Get(name)
	if store == nil {
		log.Printf("[API] Cannot find source %s", name)
		api.writeDataJSON(resp, http.StatusNotFound, "Unknown source")
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
