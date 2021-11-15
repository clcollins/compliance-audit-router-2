/*
Copyright © 2021 Red Hat, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
