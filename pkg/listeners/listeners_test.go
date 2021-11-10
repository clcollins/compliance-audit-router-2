package listeners

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestCreateListener(t *testing.T) {
	tests := []Listener{
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

	for _, tt := range tests {
		testname := fmt.Sprintf("%s %s", strings.Join(tt.Methods, "_"), tt.Path)
		t.Run(testname, func(t *testing.T) {
			l := CreateListener(tt.Path, tt.Methods, tt.HandlerFunc)
			if &tt.HandlerFunc == &l.Func {
				t.Errorf("expected %v, got %v", tt.HandlerFunc, l.Func)
			}
		})

	}
}

func TestProcessAlertHandler(t *testing.T) {
	// Example webhook payloads that might be received from the
	// alerting system (ie: Splunk)
	tests := []struct {
		name                string
		incomingWebhookBody string
		status              int
		contentType         string
		body                string
	}{
		{
			name:                "empty webhook should fail",
			incomingWebhookBody: "",
			status:              http.StatusInternalServerError,
			contentType:         "text/plain",
			body:                "failed to process webhook",
		},
		{
			name:                "testing webhook 01",
			incomingWebhookBody: "",
			status:              http.StatusOK,
			contentType:         "text/plain",
			body:                "ok",
		},
	}

	handler := http.HandlerFunc(ProcessAlertHandler)

	for _, tt := range tests {
		// Create a test incoming webhook request to the http server
		req, err := http.NewRequest(http.MethodPost, "", strings.NewReader(tt.incomingWebhookBody))
		if err != nil {
			t.Fatal(err)
		}

		// ResponseRecorder is used to store the server response
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		t.Run(tt.name, func(t *testing.T) {
			ProcessAlertHandler(recorder, req)
			// Test the returned http status code
			if status := recorder.Code; status != tt.status {
				t.Errorf("handler returned wrong status code: got %v, want %v",
					status, tt.status)
			}

			// Test the returned header Content-Type
			if contentType := recorder.Header().Get("Content-Type"); contentType != tt.contentType {
				t.Errorf("handler returned wrong Content-Type: got %v, want %v",
					contentType, tt.contentType)
			}

			// Test the returned body
			if body := recorder.Body.String(); body != string(tt.body) {
				t.Errorf("handler returned unexpected body: got %v, want %v",
					body, string(tt.body))
			}
		})
	}
}
