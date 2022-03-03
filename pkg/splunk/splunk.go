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

package splunk

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/openshift/compliance-audit-router/pkg/config"
	"github.com/openshift/compliance-audit-router/pkg/helpers"
)

// Alert describes a Splunk alert
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}
type Alert struct {
	//  TODO: convert time strings to time.Date
	FirstEvent string
	LastEvent  string
	ClusterID  string
	UserName   string
	Summary    string
	SessionID  string
	SearchID   string
}

// RetrieveSearchFromAlert parses the received webhook, and looks up the data for the alert in Splunk,
// and returns the information in an Alert struct
func RetrieveSearchFromAlert(sid string) (Alert, error) {
	var alert = Alert{}

	alert.SearchID = sid

	transport := &http.Transport{}
	// Allow insecure connections for development
	if config.AppConfig.SplunkConfig.AllowInsecure {
		log.Printf("Allowing insecure connections to Splunk")
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	// Create a new HTTP client; don't modify the default client
	splunkHttpClient := &http.Client{Transport: transport}
	req, err := http.NewRequest(
		http.MethodGet,
		getSplunkURL(config.AppConfig.SplunkConfig.Host, sid),
		http.NoBody,
	)
	if err != nil {
		return alert, err
	}

	// REALLY, Splunk?
	req.SetBasicAuth(
		config.AppConfig.SplunkConfig.Username,
		config.AppConfig.SplunkConfig.Password,
	)

	resp, err := splunkHttpClient.Do(req)
	if err != nil {
		return alert, err
	}

	// TODO:  PICK UP HERE - figure out how to parse the responses from Splunk in an alert-agnostic way
	b, err := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(b))
	os.Exit(1)

	// Process the response
	err = helpers.DecodeJSONResponseBody(resp, &alert)
	if err != nil {
		return alert, err
	}

	log.Printf("Received alert from Splunk: %+v", &alert)
	// TODO: Can we make the un-marshalling of the XML response agnostic to any specific service?  Interfaces?
	//var search Results

	os.Exit(1)
	return alert, err
}

// getSplunkURL returns the URL for a Splunk search by sid
func getSplunkURL(host, sid string) string {
	return fmt.Sprintf(
		"%s/services/search/jobs/%s/summary?output_mode=json&count=0", host, sid,
	)
}
