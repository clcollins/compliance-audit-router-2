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
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/openshift/compliance-audit-router/pkg/alerts"
)

//This file includes types for un-marshalling Splunk XML responses

// The XML response from Splunk is wrapped in a <results> tag
type Results struct {
	// XMLName xml.Name `xml:"results"`
	Results []Result `xml:"result"`
	Preview string   `xml:"preview,attr"`
}

type Result struct {
	// XMLName xml.Name `xml:"result"`
	Fields []Field `xml:"field"`
	Offset string  `xml:"offset,attr"`
}

type Field struct {
	// XMLName xml.Name `xml:"field"`
	Value Value  `xml:"value"`
	Key   string `xml:"k,attr"`
	V     string `xml:"v"`
}

type Value struct {
	// XMLName xml.Name `xml:"value"`
	Text string `xml:"text"`
}

func RetrieveSearchFromAlert(r *http.Request) (alerts.Alert, error) {
	var alert = alerts.Alert{}

	// Read the body of the request, and extract the Search ID (SID)
	// replace _ with b when we have a real webhook
	_, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return alert, err
	}

	// TODO: Extract an SID from the response
	alert.SearchID = "SID"

	// TODO: Make an HTTP request to Splunk with the SID and retrieve the events
	// For now,  load the XML response from a file
	// resp, err := SOME HTTP REQUEST HERE, using the `sid` variable

	// --- BEGIN TEMP ---
	home, err := os.UserHomeDir()
	if err != nil {
		return alert, err
	}

	xmlFile, err := os.Open(home + "/EXAMPLE_SPLUNK_WEBHOOK")
	defer xmlFile.Close()
	if err != nil {
		return alert, err
	}

	// TEMP DUMMY RESPONSE
	resp := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(xmlFile),
	}
	// --- END TEMP ---

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return alert, err
	}

	// Process the response
	// TODO: Can we make the un-marshalling of the XML response agnostic to any specific service?  Interfaces?
	var search Results

	err = xml.Unmarshal(b, &search)
	if err != nil {
		return alert, err
	}

	alert.Raw = string(b)

	for _, result := range search.Results {
		for _, field := range result.Fields {
			switch key := field.Key; key {
			case "firstEvent":
				alert.FirstEvent = field.Value.Text
			case "lastEvent":
				alert.LastEvent = field.Value.Text
			case "clusterid":
				alert.ClusterID = field.Value.Text
			case "username":
				alert.UserName = field.Value.Text
			case "elevated_summary":
				alert.Summary = field.Value.Text
			case "sessionID":
				alert.SessionID = field.Value.Text
			}

		}
	}

	return alert, err
}
