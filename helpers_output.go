package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func renderStuff(stuff interface{}, w http.ResponseWriter) {
	data, _ := json.Marshal(stuff)
	jsonString := string(data) + "\n"

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	fmt.Fprintf(w, jsonString)
}

func renderResponse(status int, message string, content interface{}, w http.ResponseWriter) {
	if content == nil {
		content = map[string]interface{}{}
	}
	renderStuff(jsonResponse{Status: status, Message: message, Content: content}, w)
}

func retrievePostJSON(r *http.Request) string { return r.PostFormValue("json") }
