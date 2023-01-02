package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "status: available")
	_, err := fmt.Fprintf(w, "environment: %s\n", app.config.env)
	if err != nil {
		return
	}
	_, err2 := fmt.Fprintf(w, "version: %s\n", version)
	if err2 != nil {
		return
	}
}
