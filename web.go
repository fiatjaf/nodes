package main

import (
	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"os"

	_ "github.com/go-cq/cq"
)

var db *sqlx.DB

func main() {
	db = sqlx.MustConnect("neo4j-cypher", os.Getenv("NEO4J_URL"))
	db = db.Unsafe()

	// middleware
	middle := interpose.New()

	// router
	router := mux.NewRouter()
	router.StrictSlash(true) // redirects '/path' to '/path/'
	middle.UseHandler(router)
	// ~

	// > normal pages and index
	router.HandleFunc("/r/", CreateRelationship).Methods("POST")
	router.HandleFunc("/r/", ViewRelationships).Methods("GET")
	// ~

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Print("listening...")
	http.ListenAndServe(":"+port, middle)
}
