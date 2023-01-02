package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	var env = envelope{
		Id:       "1",
		TypeData: "healthCheck",
		Attributes: map[string]any{
			"environment": app.config.env,
			"version":     version,
		},
	}

	var err = app.writeJSON(w, http.StatusOK, env, nil)

	if err != nil {
		app.logger.Println(err)
		http.Error(w, "There is a problem in decoding data", http.StatusInternalServerError)
	}
}
