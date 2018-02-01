package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"opencoredata.org/ocdSearch/search"
)

// MyServer is the Gorilla mux router struct
type MyServer struct {
	r *mux.Router
}

func main() {
	searchroute := mux.NewRouter()
	searchroute.HandleFunc("/search", search.DoSearch)
	http.Handle("/search", searchroute)

	imageRouter := mux.NewRouter()
	imageRouter.PathPrefix("/search/images/").Handler(http.StripPrefix("/search/images/", http.FileServer(http.Dir("./images"))))
	http.Handle("/search/images/", &MyServer{imageRouter})

	cssRouter := mux.NewRouter()
	cssRouter.PathPrefix("/search/css/").Handler(http.StripPrefix("/search/css/", http.FileServer(http.Dir("./css"))))
	http.Handle("/search/css/", &MyServer{cssRouter})

	log.Printf("About to listen on 9900. Go to http://127.0.0.1:9900/")

	err := http.ListenAndServe(":9900", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *MyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// Let the Gorilla work
	s.r.ServeHTTP(rw, req)
}

func addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fn(w, r)
	}
}
