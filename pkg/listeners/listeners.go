package listeners

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openshift/compliance-audit-router/pkg/config"
	"github.com/openshift/compliance-audit-router/pkg/processor"
)

// Handler defines an HTTP route handler
type Handler struct {
	Route func(r *mux.Route)
	Func  http.HandlerFunc
}

type Listener struct {
	Path        string
	Methods     []string
	HandlerFunc http.HandlerFunc
}

var Listeners = []Listener{
	{
		Path:        "/readyz",
		Methods:     []string{http.MethodGet},
		HandlerFunc: http.HandlerFunc(LogAndRespondOKHandler),
	},
	{
		Path:        "/healthz",
		Methods:     []string{http.MethodGet},
		HandlerFunc: http.HandlerFunc(LogAndRespondOKHandler),
	},
	{
		Path:        "/api/v1/alert",
		Methods:     []string{http.MethodPost},
		HandlerFunc: http.HandlerFunc(ProcessAlertHandler),
	},
}

func CreateListener(path string, methods []string, handlerFunc http.HandlerFunc) Handler {
	if config.AppConfig.Verbose {
		log.Println("enabling endpoint", path, methods)
	}

	return Handler{
		Route: func(r *mux.Route) {
			r.Path(path).Methods(methods...)
		},
		Func: handlerFunc,
	}
}

func (h Handler) AddRoute(r *mux.Router) {
	h.Route(r.NewRoute().HandlerFunc(h.Func))
}

func RequestLogger(r *http.Request) {
	log.Println(r.RemoteAddr, r.Method, r.RequestURI)
	if config.AppConfig.Verbose {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		log.Println(string(b))
	}
}

func LogAndRespondOKHandler(w http.ResponseWriter, r *http.Request) {
	RequestLogger(r)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

func ProcessAlertHandler(w http.ResponseWriter, r *http.Request) {
	RequestLogger(r)
	err := processor.ProcessAlert(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Failed to process webhook"))
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Processed"))
	}
}
