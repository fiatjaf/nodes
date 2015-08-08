package main

import (
	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/jmcvetta/neoism"
	"log"
	"net/http"
	"os"

	"nodes/handlers"
)

var db *neoism.Database

func main() {
	var err error
	db, err = neoism.Connect(os.Getenv("NEO4J_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// middleware
	middle := interpose.New()

	// router
	router := mux.NewRouter()
	router.StrictSlash(true) // redirects '/path' to '/path/'
	middle.UseHandler(router)
	// ~

	// > routes
	router.HandleFunc("/rel/", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateRelationship(db, w, r)
	}).Methods("POST")
	router.HandleFunc("/eql/", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateEquality(db, w, r)
	}).Methods("POST")
	router.HandleFunc("/rels.svg", func(w http.ResponseWriter, r *http.Request) {
		handlers.ViewRelationships(db, w, r)
	}).Methods("GET")
	// ~

	// static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./site/")))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	// ~

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Print("listening...")
	http.ListenAndServe(":"+port, middle)
}
