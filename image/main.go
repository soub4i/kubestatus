package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Service struct {
	Name   string
	URI    string
	Status bool
}

type HTML struct {
	Services []Service
}

func ping(s string) bool {
	r, e := http.Head(s)
	return e == nil && r.StatusCode == 200
}

func main() {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		services := os.Getenv("services")

		var data = make([]Service, 0, 50)
		for _, e := range strings.Split(services, ";") {
			pair := strings.SplitN(e, "=", 2)
			if len(pair) == 2 {
				fmt.Println("Health checking: " + pair[1])
				chunk := strings.SplitN(pair[1], ":", 2)
				status := false
				if len(chunk) == 2 {
					status = ping("http://" + chunk[0] + chunk[1])
				} else {
					status = ping("http://" + pair[1])
				}
				data = append(data, Service{Name: pair[0], Status: status})
			}
		}
		json.NewEncoder(w).Encode(data)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Alive")
	})

	http.ListenAndServe(":8080", nil)
}
