package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/openshift/compliance-audit-router/pkg/infoLog"
	"github.com/openshift/compliance-audit-router/pkg/listeners"
)

const (
	listenPort = ":8080"
)

var verbose bool

var l = []listeners.Listener{
	{
		Path:        "/readyz",
		Methods:     []string{http.MethodGet},
		HandlerFunc: http.HandlerFunc(listeners.LogAndRespondOKHandler),
	},
	{
		Path:        "/healthz",
		Methods:     []string{http.MethodGet},
		HandlerFunc: http.HandlerFunc(listeners.LogAndRespondOKHandler),
	},
	{
		Path:        "/api/v1/alert",
		Methods:     []string{http.MethodPost},
		HandlerFunc: http.HandlerFunc(listeners.ProcessAlertHandler),
	},
}

func init() {
	flag.BoolVar(&verbose, "v", false, "Enable verbose logging")
}

func main() {
	flag.Parse()

	r := mux.NewRouter()

	for _, listener := range l {
		listeners.CreateListener(listener.Path, listener.Methods, listener.HandlerFunc, verbose).AddRoute(r)
	}

	infoLog.Logger.Printf("Listening on %s", listenPort)
	log.Fatal(http.ListenAndServe(listenPort, r))
}
