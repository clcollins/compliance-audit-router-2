package listeners

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openshift/compliance-audit-router/pkg/config"
	"github.com/openshift/compliance-audit-router/pkg/jira"
	"github.com/openshift/compliance-audit-router/pkg/ldap"
	"github.com/openshift/compliance-audit-router/pkg/splunk"
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

func LogAndRespondOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

func ProcessAlertHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the alert search results

	var err error
	searchResults, err := splunk.RetrieveSearchFromAlert(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("failed alert lookup"))
		return
	}

	//user, manager, err := ldap.LookupUser(searchResults.UserName)
	user, manager, err := ldap.LookupUser("chcollin")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("failed ldap lookup"))
		return
	}

	err = jira.CreateTicket(user, manager, searchResults)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("failed ticket creation"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ok"))
	return
}
