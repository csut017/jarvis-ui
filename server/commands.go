package main

type command struct {
	Name     string `json:"name"`
	Action   string `json:"action"`
	Duration *int   `json:"duration"`
}
