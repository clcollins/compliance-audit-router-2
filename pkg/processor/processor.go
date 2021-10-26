package processor

import (
	"net/http"

	"github.com/openshift/compliance-audit-router/pkg/jira"
	"github.com/openshift/compliance-audit-router/pkg/ldap"
	"github.com/openshift/compliance-audit-router/pkg/splunk"
)

// ProcessAlert receives an alert webhook and handles the logic of looking up the alert and handling it
func ProcessAlert(r *http.Request) error {
	// Retrieve the alert search results
	searchResults, err := splunk.RetrieveSearchFromAlert(r)
	if err != nil {
		return err
	}

	user, manager, err := ldap.LookupUser(searchResults.UserName)
	if err != nil {
		return err
	}

	err = jira.CreateTicket(
		user,
		manager,
		searchResults.FirstEvent,
		searchResults.ClusterID,
		searchResults.Summary,
		searchResults.SessionID,
		searchResults.Raw,
	)

	if err != nil {
		return err
	}

	return nil
}
