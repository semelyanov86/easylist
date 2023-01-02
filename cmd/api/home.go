package main

import (
	"log"
	"net/http"
	"text/template"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ts, err := template.ParseFiles("./ui/html/home.page.html")
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
