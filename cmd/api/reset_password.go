package main

import (
	"log"
	"net/http"
	"text/template"
)

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("reset_token")

	if token == "" {
		http.Error(w, "No token provided", http.StatusUnprocessableEntity)
		return
	}

	ts, err := template.ParseFiles("./ui/html/reset.page.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = ts.Execute(w, token)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
