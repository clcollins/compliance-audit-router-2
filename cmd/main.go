package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/openshift/compliance-audit-router/pkg/config"
	"github.com/openshift/compliance-audit-router/pkg/listeners"
)

func main() {
	r := mux.NewRouter()

	for _, listener := range listeners.Listeners {
		listeners.CreateListener(listener.Path, listener.Methods, listener.HandlerFunc).AddRoute(r)
	}

	portString := ":" + fmt.Sprint(config.AppConfig.ListenPort)
	log.Printf("Listening on %s", portString)
	log.Fatal(http.ListenAndServe(portString, r))
}
