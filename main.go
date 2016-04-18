package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"opencoredata.org/ocdSearch/handler"
)

// +build !appengine

var bindAddr = flag.String("addr", ":8080", "http listen address")

func main() {
	// Route for common files
	rcommon := mux.NewRouter()
	rcommon.PathPrefix("/common/").Handler(http.StripPrefix("/common/", http.FileServer(http.Dir("./static"))))

	// Route for main / handle
	hndlroute := mux.NewRouter()
	hndlroute.HandleFunc("/", handler.DoSearch)

	// Server mux
	serverMuxA := http.NewServeMux()
	serverMuxA.Handle("/", hndlroute)
	serverMuxA.Handle("/common/", rcommon)

	go func() {
		http.ListenAndServe("localhost:8082", serverMuxA)
	}()
	log.Printf("Listening for HTML  on %v", 8082)

	// Start the Bleve search API services running
	flag.Parse()
	log.Printf("Listening on %v", *bindAddr)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}
