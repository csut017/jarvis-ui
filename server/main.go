package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/tarm/serial"
)

func main() {
	log.Printf("[Server] Starting...")

	var (
		configPath = flag.String("config", "config.json", "The configuration file to use")
		port       = flag.String("port", "80", "port to serve on")
		monitors   = &monitorStore{}
		weather    = &weatherService{}
	)
	flag.Parse()

	log.Printf("[Main] Reading configuration from %s", *configPath)
	config, err := readConfiguration(*configPath)
	if err != nil {
		log.Fatalf("[Main] Unable to read configuration: %v", err)
	}

	log.Printf("[Main] Starting data store")
	data := &dataStore{}
	dataChan := data.Initialise()
	data.Start()

	if config.Weather != nil {
		log.Printf("[Main] Starting weather service")
		weather.Start(config.Weather)
	}

	log.Printf("[Main] Initialising webserver")
	addr := ":" + *port
	api, srv := initialiseWebServer(addr, data, monitors, weather, config)
	out := make(chan *monitorResult)
	go handleResult(out, api)

	log.Printf("[Main] Starting monitors")
	for _, sensor := range config.Sources {
		if sensor.IsEnabled {
			log.Printf("[Main] Starting monitor %s", sensor.Name)
			sourceConfig := &serial.Config{Name: sensor.Port, Baud: 9600}
			mon := &monitor{}
			err := mon.Start(sourceConfig, sensor.Name)
			if err != nil {
				log.Fatalf("[Main] Unable to start monitor %s: %v", sensor.Name, err)
			}

			mon.AddListener(out)
			mon.AddListener(dataChan)
			monitors.Add(sensor.Name, mon)
		} else {
			log.Printf("[Main] Skipping monitor %s - disabled", sensor.Name)
		}
	}

	log.Printf("[Main] Starting webserver")
	api.start()
	go func() {
		log.Printf("[WebServer] Listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Printf("[WebServer] Stopped")
			} else {
				log.Printf("[WebServer] Unable to start: %v", err)
			}
		}
	}()
	waitForShutdown(srv)

	log.Printf("[Main] Stopping monitors")
	for _, mon := range *monitors {
		mon.Stop()
		if err = mon.LastError(); err != nil {
			log.Fatalf("[Main] Unable to stop monitor %s: %v", mon.Name(), err)
		}
	}
	weather.Stop(time.Second * 5)
	close(out)
}
func handleResult(input <-chan *monitorResult, srv *webAPI) {
	for {
		result, open := <-input
		if open {
			log.Printf("[Main] Received %+v", result)
			srv.hub.sendResult(result)
		} else {
			return
		}
	}
}

func initialiseWebServer(addr string, data *dataStore, monitors *monitorStore, weather *weatherService, config *appConfiguration) (*webAPI, *http.Server) {
	rootMiddleware := interpose.New()

	rootRouter := mux.NewRouter()
	rootRouter.PathPrefix("/s/css").Handler(http.StripPrefix("/s/css", http.FileServer(http.Dir(filepath.Join(config.StaticPath, "css")))))
	rootRouter.PathPrefix("/s/media").Handler(http.StripPrefix("/s/media", http.FileServer(http.Dir(filepath.Join(config.StaticPath, "media")))))
	rootRouter.PathPrefix("/s/js").Handler(http.StripPrefix("/s/js", http.FileServer(http.Dir(filepath.Join(config.StaticPath, "js")))))
	rootRouter.HandleFunc("/favicon.ico", func(resp http.ResponseWriter, req *http.Request) {
		http.ServeFile(resp, req, filepath.Join(config.StaticPath, "media/favicon.ico"))
	})

	rootMiddleware.Use(logRequestsMiddleware)
	rootMiddleware.UseHandler(rootRouter)

	api, err := newWebAPI(addr, data, monitors, weather, config)
	if err != nil {
		log.Fatalf("[Main] Unable to initialise API: %v", err)
	}

	apiRouter := api.Router
	apiMiddleware := interpose.New()
	apiMiddleware.Use(setOriginRequestAndHeadersMiddleware)
	apiMiddleware.UseHandler(apiRouter)
	rootRouter.PathPrefix("/api").Handler(apiMiddleware)

	fileServeRouter := formatFileName(http.FileServer(http.Dir(filepath.Join(config.StaticPath, "html"))))
	fileServeMiddleware := interpose.New()
	fileServeMiddleware.Use(disableCaching)
	fileServeMiddleware.UseHandler(fileServeRouter)
	rootRouter.PathPrefix("/").Handler(fileServeMiddleware)

	srv := &http.Server{
		Addr:         addr,
		Handler:      rootMiddleware,
		IdleTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 20,
		ReadTimeout:  time.Second * 20,
	}
	return api, srv
}

func setOriginRequestAndHeadersMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if origin := req.Header.Get("Origin"); origin != "" {
			resp.Header().Set("Access-Control-Allow-Origin", origin)
			resp.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
			resp.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		if req.Method == "OPTIONS" {
			return
		}

		handler.ServeHTTP(resp, req)
	})
}

func disableCaching(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		resp.Header().Set("Pragma", "no-cache")
		resp.Header().Set("Expires", "0")
		handler.ServeHTTP(resp, req)
	})
}

func waitForShutdown(srv *http.Server) {
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			log.Printf("[WebServer] Stopping")
			ctx, f := context.WithTimeout(context.Background(), 15*time.Second)
			if err := srv.Shutdown(ctx); err != nil {
				log.Println("[WebServer] Unable to shutdown server gracefully: " + err.Error())
				srv.Close()
			}
			f()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}

func formatFileName(h http.Handler) http.Handler {
	stripper := http.StripPrefix("/", h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if !strings.HasSuffix(p, ".html") && (p != "/") {
			p = p + ".html"
			log.Printf("[Main] Mapped URL to %s", p)
		}
		r.URL.Path = p
		stripper.ServeHTTP(w, r)
	})
}
