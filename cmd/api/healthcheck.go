package main

import (
	"net/http"
)

type healthCheck struct {
	ID          int64  `jsonapi:"primary,health_check"`
	Environment string `jsonapi:"attr,environment"`
	Version     string `jsonapi:"attr,version"`
	Status      string `jsonapi:"attr,status"`
}

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	env := healthCheck{
		ID:          1,
		Environment: app.config.Env,
		Version:     version,
		Status:      "available",
	}

	var err = app.writeJSON(w, http.StatusOK, &env, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
