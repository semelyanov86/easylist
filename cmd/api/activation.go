package main

import (
	"log"
	"net/http"
	"text/template"
)

func activation(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/activate" {
		http.NotFound(w, r)
		return
	}
	ts, err := template.ParseFiles("./ui/html/activation.page.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal server error", 500)
		return
	}
	err = ts.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal server error", 500)
	}
}
