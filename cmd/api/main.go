package main

import (
	"net/http"
	"wikidist/pkg/api"
	"log"
)

func main(){
    http.HandleFunc("/", api.TestHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}