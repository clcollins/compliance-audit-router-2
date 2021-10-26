package listeners

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/openshift/compliance-audit-router/pkg/alerts"
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
	log.Println(r.RemoteAddr, r.Method, r.RequestURI)
	// TODO: remove this debug
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(b))
}

func LogAndRespondOKHandler(w http.ResponseWriter, r *http.Request) {
	RequestLogger(r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ProcessAlertHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("got here")
	RequestLogger(r)
	resp, err := alerts.ProcessAlert(r)
	if err != nil {
		log.Println(err)
	}

	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(resp.Body))
}
