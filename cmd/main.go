package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/openshift/compliance-audit-router/pkg/config"
	"github.com/openshift/compliance-audit-router/pkg/listeners"
)

var portString = ":" + fmt.Sprint(config.AppConfig.ListenPort)

func main() {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	for _, listener := range listeners.Listeners {
		listeners.CreateListener(listener.Path, listener.Methods, listener.HandlerFunc).AddRoute(r)
	}

	log.Printf("Listening on %s", portString)
	log.Fatal(http.ListenAndServe(portString, r))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, next)
}
