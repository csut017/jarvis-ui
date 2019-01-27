package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

type monitorResult struct {
	Source    string               `json:"source"`
	TimeStamp string               `json:"time"`
	Counter   int64                `json:"count"`
	Values    []monitorResultValue `json:"values"`
}

type monitorResultValue struct {
	Name  string  `json:"name"`
	Value float32 `json:"value"`
}

type monitorListener chan<- *monitorResult

type monitorStore map[string]*monitor

func (store monitorStore) Get(name string) *monitor {
	return store[name]
}

func (store monitorStore) Add(name string, mon *monitor) {
	store[name] = mon
}

type monitor struct {
	name         string
	conn         *serial.Port
	running      bool
	stopping     bool
	stopResult   chan int
	lastError    error
	listeners    map[monitorListener]bool
	counter      int64
	inputValues  []string
	outputValues []string
	mux          sync.Mutex
}

func (mon *monitor) AddListener(listener monitorListener) {
	if mon.listeners == nil {
		mon.listeners = map[monitorListener]bool{}
	}
	mon.listeners[listener] = true
}

func (mon *monitor) RemoveListener(listener monitorListener) {
	if mon.listeners == nil {
		return
	}
	if mon.listeners[listener] {
		delete(mon.listeners, listener)
	}
}

func (mon *monitor) Start(config *serial.Config, name string) error {
	if mon.running {
		return fmt.Errorf("Monitor is already running")
	}

	mon.name = name
	config.ReadTimeout = time.Second
	conn, err := serial.OpenPort(config)
	if err != nil {
		return fmt.Errorf("Unable to start monitor: %v", err)
	}

	mon.stopResult = make(chan int)
	mon.conn = conn
	mon.stopping = false
	mon.lastError = nil
	go mon.run()

	return nil
}

func (mon *monitor) Stop() error {
	if mon.running {
		return fmt.Errorf("Monitor is not running")
	}

	mon.stopping = true
	_ = <-mon.stopResult
	return nil
}

func (mon *monitor) Name() string {
	return mon.name
}

func (mon *monitor) IsRunning() bool {
	return mon.running
}

func (mon *monitor) LastError() error {
	return mon.lastError
}

func (mon *monitor) InputTypes() []string {
	mon.mux.Lock()
	defer mon.mux.Unlock()
	return mon.inputValues
}

func (mon *monitor) GetSensorPosition(name string) int {
	mon.mux.Lock()
	defer mon.mux.Unlock()
	for index, sensor := range mon.inputValues {
		if sensor == name {
			return index
		}
	}
	return -1
}

func (mon *monitor) OutputTypes() []string {
	mon.mux.Lock()
	defer mon.mux.Unlock()
	return mon.outputValues
}

func (mon *monitor) SendCommand(cmd *command) error {
	outputNumber := -1
	mon.mux.Lock()
	for loop := 0; loop < len(mon.outputValues); loop++ {
		if mon.outputValues[loop] == cmd.Name {
			outputNumber = loop
			break
		}
	}
	mon.mux.Unlock()

	if outputNumber < 0 {
		return fmt.Errorf("Unable to find effector '%s'", cmd.Name)
	}

	action := " "
	switch cmd.Action {
	case "on":
		action = "+"
	case "off":
		action = "-"

	default:
		return fmt.Errorf("Unknown action '%s'", cmd.Action)
	}
	msg := ""
	if cmd.Duration != nil && *cmd.Duration > 0 {
		msg = fmt.Sprintf("C:%d%s%d", outputNumber, action, *cmd.Duration)
	} else {
		msg = fmt.Sprintf("C:%d%s", outputNumber, action)
	}

	log.Printf("[Monitor] Sending '%s' to %s", msg, mon.name)
	n, err := mon.conn.Write([]byte(msg + "\n"))
	if err != nil {
		log.Printf("[Monitor] Error sending '%s' to %s: %v", msg, mon.name, err)
		return fmt.Errorf("Unable to connect to source")
	} else {
		log.Printf("[Monitor] Sent %d bytes to %s", n, mon.name)
	}

	return nil
}

func (mon *monitor) run() {
	defer mon.conn.Close()
	log.Printf("[Monitor] Monitor %s started", mon.name)

	mon.running = true
	out := &bytes.Buffer{}
	for !mon.stopping {
		buf := make([]byte, 1024)
		for loop := 0; loop < 10; loop++ {
			n, err := mon.conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					mon.stopping = true
					log.Printf("[Monitor] Read error on %s: %v", mon.name, err)
					mon.lastError = err
				}
			} else if n > 0 {
				log.Printf("[Monitor] Received %d bytes from %s", n, mon.name)
			}
			for loop := 0; loop < n; loop++ {
				char := buf[loop]
				switch char {
				case '\r':
					// Ignore carriage returns

				case '\n':
					rawData := out.String()
					if len(rawData) > 0 {
						msgData := strings.Split(rawData[2:], ",")
						switch rawData[0] {
						case 'O':
							mon.loadInputTypes(msgData)

						case 'I':
							mon.loadOutputTypes(msgData)

						case 'D':
							mon.readData(msgData)

						case 'C':
							mon.handleCommand(msgData)

						default:
							log.Printf("[Monitor] Received unknown input '%s' from %s", rawData, mon.name)
						}
						out.Reset()
					}
				default:
					out.WriteByte(char)
				}
			}
		}
	}

	mon.running = false
	close(mon.stopResult)
	log.Printf("[Monitor] Monitor %s finished", mon.name)
}

func (mon *monitor) loadOutputTypes(values []string) {
	log.Printf("[Monitor] Received output types %v from %s", values, mon.name)
	mon.mux.Lock()
	mon.outputValues = values
	mon.mux.Unlock()
}

func (mon *monitor) loadInputTypes(values []string) {
	log.Printf("[Monitor] Received input types %v from %s", values, mon.name)
	mon.mux.Lock()
	mon.inputValues = values
	mon.mux.Unlock()
}

func (mon *monitor) readData(values []string) {
	log.Printf("[Monitor] Received values %v from %s", values, mon.name)
	result := &monitorResult{
		Source:    mon.name,
		TimeStamp: time.Now().Format(time.RFC3339),
		Counter:   mon.counter,
		Values:    make([]monitorResultValue, len(mon.inputValues)),
	}

	for loop := 0; loop < len(mon.inputValues); loop++ {
		result.Values[loop] = monitorResultValue{
			Name:  mon.inputValues[loop],
			Value: parseFloat(values[loop]),
		}
	}

	mon.counter++
	if mon.listeners != nil {
		for listener := range mon.listeners {
			listener <- result
		}
	}
}

func (mon *monitor) handleCommand(values []string) {
	log.Printf("[Monitor] Received command %v from %s", values, mon.name)
}

func parseFloat(value string) float32 {
	val, err := strconv.ParseFloat(value, 32)
	if err != nil {
		log.Printf("[Monitor] WARNING: Unable to parse '%s' as a float", value)
	}
	return float32(val)
}
