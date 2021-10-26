package processor

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/openshift/compliance-audit-router/pkg/alerts"
	"github.com/openshift/compliance-audit-router/pkg/jira"
	"github.com/openshift/compliance-audit-router/pkg/ldap"
	"github.com/openshift/compliance-audit-router/pkg/splunk"
)

var failedResponse = alerts.ProcessResponse{StatusCode: http.StatusInternalServerError, Body: "Failed processing webhook"}

// ProcessAlert receives an alert webhook and handles the logic of looking up the alert and handling it
func ProcessAlert(r *http.Request) (alerts.ProcessResponse, error) {
	var sid string

	// Read the body of the request, and extract the Search ID (SID)
	// replace _ with b when we have a real webhook
	_, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return failedResponse, err
	}

	// TODO: Extract an SID from the response
	sid = "foo_bar_baz"
	fmt.Println("FAKE SID: ", sid)

	// TODO: Make an HTTP request to Splunk with the SID and retrieve the events
	// For now,  load the XML response from a file
	// resp, err := SOME HTTP REQUEST HERE, using the `sid` variable

	// --- BEGIN TEMP ---
	home, err := os.UserHomeDir()
	if err != nil {
		return failedResponse, err
	}

	xmlFile, err := os.Open(home + "/EXAMPLE_SPLUNK_WEBHOOK")
	defer xmlFile.Close()
	if err != nil {
		return failedResponse, err
	}

	// TEMP DUMMY RESPONSE
	resp := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(xmlFile),
	}
	// --- END TEMP ---

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return failedResponse, err
	}

	// Process the response
	// TODO: Can we make the un-marshalling of the XML response agnostic to any specific service?  Interfaces?
	var search splunk.Results

	err = xml.Unmarshal(b, &search)
	if err != nil {
		return failedResponse, err
	}

	var alert = alerts.Alert{}
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

	user, manager, err := ldap.LookupUser(alert.UserName)
	err = jira.CreateTicket(
		user,
		manager,
		alert.FirstEvent,
		alert.ClusterID,
		alert.Summary,
		alert.SessionID,
		alert.Raw,
	)

	return alerts.ProcessResponse{StatusCode: http.StatusOK, Body: "Processed"}, err
}
