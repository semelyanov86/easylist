package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	var env = envelope{
		Id:       1,
		TypeData: "healthCheck",
		Attributes: map[string]any{
			"environment": app.config.Env,
			"version":     version,
			"status":      "available",
		},
	}

	var err = app.writeJSON(w, http.StatusOK, env, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
