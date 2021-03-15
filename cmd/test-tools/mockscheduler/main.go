/*
Copyright 2021 Adevinta
*/

package main

import (
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe(":8082", nil))
}
