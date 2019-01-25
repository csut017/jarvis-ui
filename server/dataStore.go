package main

import "fmt"

const storeSize = 1000

type sourceDataStore struct {
	Name     string
	items    []monitorResult
	position int
}

func newSourceDataStore(name string) *sourceDataStore {
	item := &sourceDataStore{
		items:    make([]monitorResult, 2*storeSize),
		Name:     name,
		position: -1,
	}
	return item
}

func (store *sourceDataStore) Add(result *monitorResult) {
	store.position++
	if store.position >= 2*storeSize {
		// Need to re-position the array
		for loop := 0; loop < storeSize-1; loop++ {
			store.items[loop] = store.items[loop+storeSize+1]
		}
		store.position = storeSize
	}
	store.items[store.position] = *result
}

func (store *sourceDataStore) Get() *[]monitorResult {
	start := store.position - storeSize
	if start < 0 {
		start = 0
	}
	size := store.position - start + 1
	out := make([]monitorResult, size)
	for loop := 0; loop < size; loop++ {
		out[loop] = store.items[loop+start]
	}
	return &out
}

type dataStore struct {
	sources    map[string]*sourceDataStore
	input      chan *monitorResult
	stopSignal chan int
	stopResult chan int
	running    bool
}

func (store *dataStore) Initialise() monitorListener {
	store.sources = make(map[string]*sourceDataStore)
	store.input = make(chan *monitorResult)
	return store.input
}

func (store *dataStore) Start() error {
	if store.running {
		return fmt.Errorf("Data store is already running")
	}

	store.stopSignal = make(chan int)
	store.stopResult = make(chan int)
	go store.run()
	return nil
}

func (store *dataStore) Stop() error {
	if store.running {
		return fmt.Errorf("Data store is not running")
	}

	store.stopSignal <- 1
	_ = <-store.stopResult
	return nil
}

func (store *dataStore) IsRunning() bool {
	return store.running
}

func (store *dataStore) GetItems(name string) *[]monitorResult {
	source, ok := store.sources[name]
	if !ok {
		return &[]monitorResult{}
	}
	return source.Get()
}

func (store *dataStore) run() {
	store.running = true
	running := true
	for running {
		select {
		case _ = <-store.stopSignal:
			running = false

		case result, ok := <-store.input:
			if ok {
				name := result.Source
				source, ok := store.sources[name]
				if !ok {
					source = newSourceDataStore(name)
					store.sources[name] = source
				}
				source.Add(result)

			} else {
				running = false
			}
		}
	}
	store.running = false
	close(store.stopResult)
}
