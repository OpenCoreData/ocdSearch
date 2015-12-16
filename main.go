package main

import (
	"flag"
	"log"
	"net/http"
)

// +build !appengine

var bindAddr = flag.String("addr", ":8080", "http listen address")

func main() {

	// HTTP interface
	// Simple sevice for some static pages about the glservice
	serverMuxA := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	serverMuxA.Handle("/", fs)

	go func() {
		http.ListenAndServe("localhost:8082", serverMuxA)
	}()
	log.Printf("Listening for HTML  on %v", 8082)

	// Bleve search API
	flag.Parse()
	log.Printf("Listening on %v", *bindAddr)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}
