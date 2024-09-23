package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	handleCategoryRequests(router)
	handleSpeciesRequests(router)
	handleAnimalRequests(router)
	log.Fatal(http.ListenAndServe(":10000", router))
}
