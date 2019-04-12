package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	//Router setup
	//Initialize
	router := mux.NewRouter()
	//Route handlers (endpoints)
	router.HandleFunc("/804583589/all", getAll).Methods("GET")
	router.HandleFunc("/804583589/status", getStatus).Methods("GET")
	//Run server
	if err := http.ListenAndServe(":9290", router); err != nil {
		fmt.Println(err)
	}
}

//Gets JSON representation of the entire table
func getAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(playlist)
}

//Gets status of the table
func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(playlist)
}
