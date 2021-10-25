package listeners

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/openshift/compliance-audit-router/pkg/alerts"
	"github.com/openshift/compliance-audit-router/pkg/infoLog"
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

func CreateListener(path string, methods []string, handlerFunc http.HandlerFunc, verbose bool) Handler {
	if verbose {
		log.Println("enabling endpoint", path, methods)
	}

	_, err := url.ParseRequestURI(path)
	if err != nil {
		log.Panic(err)
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
	infoLog.Logger.Println(r.RemoteAddr, r.Method, r.RequestURI)
}

func LogAndRespondOKHandler(w http.ResponseWriter, r *http.Request) {
	RequestLogger(r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ProcessAlertHandler(w http.ResponseWriter, r *http.Request) {
	RequestLogger(r)
	resp, err := alerts.ProcessAlert(r)
	if err != nil {
		log.Println(err)
	}

	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(resp.Body))
}
