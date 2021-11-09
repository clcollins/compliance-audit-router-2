package listeners

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestListenerURIs(t *testing.T) {
	for _, l := range Listeners {
		_, err := url.ParseRequestURI(l.Path)
		if err != nil {
			t.Errorf("%s is not a valid url", l.Path)
		}
	}
}

var testListeners = []Listener{
	{
		Path:    "/testListener01",
		Methods: []string{http.MethodGet},
		HandlerFunc: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test"))
			},
		),
	},
	{
		Path:    "/testListener02",
		Methods: []string{http.MethodGet, http.MethodPost},
		HandlerFunc: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test"))
			},
		),
	},
}

func TestCreateListener(t *testing.T) {
	for _, tt := range testListeners {
		testname := fmt.Sprintf("%s %s", strings.Join(tt.Methods, "/"), tt.Path)
		t.Run(testname, func(t *testing.T) {
			l := CreateListener(tt.Path, tt.Methods, tt.HandlerFunc)
			if &tt.HandlerFunc == &l.Func {
				t.Errorf("expected %v, got %v", tt.HandlerFunc, l.Func)
			}
		})

	}
}
