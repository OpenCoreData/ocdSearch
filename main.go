package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"opencoredata.org/ocdSearch/handler"
)

var bindAddr = flag.String("addr", ":9800", "http listen address")

func main() {
	// Route for common files
	rcommon := mux.NewRouter()
	rcommon.PathPrefix("/ocdsearchcommon/").Handler(http.StripPrefix("/ocdsearchcommon/", http.FileServer(http.Dir("./static"))))

	// Route for main / handle
	hndlroute := mux.NewRouter()
	hndlroute.HandleFunc("/search", handler.DoSearch)

	// Server mux
	serverMuxA := http.NewServeMux()
	serverMuxA.Handle("/search", hndlroute)
	serverMuxA.Handle("/ocdsearchcommon/", rcommon)

	go func() {
		http.ListenAndServe(":9802", serverMuxA)
	}()
	log.Printf("Listening for HTTP/HTML calls on %v", 9802)

	// Start the Bleve search API services running on port 8080
	flag.Parse()
	log.Printf("Listening for HTTP/API calls on %v", *bindAddr)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}
