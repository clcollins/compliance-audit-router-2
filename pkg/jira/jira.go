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

package jira

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/openshift/compliance-audit-router/pkg/config"
	"github.com/openshift/compliance-audit-router/pkg/splunk"
)

const managedLabel string = "compliance-audit-router-managed"

func CreateTicket(user, manager string, searchResults splunk.SearchResult) error {
	return nil
}

// func Run is a wrapper for initial jira-go testing
func Run() {
	// Auth
	transport := jira.PATAuthTransport{
		Token: config.AppConfig.JiraConfig.Token,
	}

	c, err := jira.NewClient(transport.Client(), config.AppConfig.JiraConfig.Host)
	if err != nil {
		fmt.Print(err)
	}

	issues, err := GetAllIssues(c, config.AppConfig.JiraConfig.Query)
	if err != nil {
		fmt.Print(err)
	}

	for _, issue := range issues {
		doSomethingWithIssue(c, &issue)
	}
}

func GetAllIssues(client *jira.Client, searchString string) ([]jira.Issue, error) {
	last := 0
	var issues []jira.Issue
	for {
		opt := &jira.SearchOptions{
			MaxResults: 100,
			StartAt:    last,
			Fields:     []string{"created", "summary", "assignee", "labels", "attachment"},
		}

		chunk, resp, err := client.Issue.Search(searchString, opt)
		if err != nil {
			return nil, err
		}

		total := resp.Total
		if issues == nil {
			issues = make([]jira.Issue, 0, total)
		}

		issues = append(issues, chunk...)

		last = resp.StartAt + len(chunk)
		if last >= total {
			return issues, nil
		}
	}
}

func doSomethingWithIssue(client *jira.Client, issue *jira.Issue) error {
	// Unassigned issues don't have an Assignee object to grab the name from
	if issue.Fields.Assignee == nil {
		fmt.Printf("\t%s\tUNASSIGNED\t%s\n", issue.Key, issue.Fields.Summary)
	} else {
		fmt.Printf("\t%s\t%s\t%s\n", issue.Key, issue.Fields.Assignee.DisplayName, issue.Fields.Summary)
	}

	if !issueHasLabel(issue, managedLabel) {
		log.Printf("Adding compliance label to issue %s\n", issue.Key)

		labels := map[string]interface{}{
			"fields": map[string]interface{}{
				"labels": append(issue.Fields.Labels, managedLabel),
			},
		}

		_, err := client.Issue.UpdateIssue(issue.Key, labels)
		if err != nil {
			return err
		}
	}

	var alert_data []map[string]string

	for _, attachment := range issue.Fields.Attachments {
		if validateAttachment(issue.Fields.Summary, issue.Fields.Created, attachment) {
			resp, err := client.Issue.DownloadAttachment(attachment.ID)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("Error retrieving attachment: %s", resp.Status)
			}

			csv_reader := csv.NewReader(resp.Body)
			csv_reader.Comma = ','

			alert_data, err = convertCSVToMap(csv_reader)
			if err != nil {
				return err
			}

			fmt.Println(alert_data)
		}
	}

	fmt.Printf("\tUser: %s\n", alert_data[0]["User"])
	// fmt.Printf("\tBackplaneID: %s\n", alert_data[0]["BackplaneID"])
	// fmt.Printf("\tCluster: %s\n", alert_data[0]["clusterid"])

	return nil
}

func convertCSVToMap(reader *csv.Reader) (mapData []map[string]string, err error) {
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	header := []string{} // holds first row (header)
	for lineNum, record := range data {

		// for first row, build the header slice
		if lineNum == 0 {
			for i := 0; i < len(record); i++ {
				header = append(header, strings.TrimSpace(record[i]))
			}
		} else {
			// for each cell, map[string]string k=header v=value
			line := map[string]string{}
			for i := 0; i < len(record); i++ {
				line[header[i]] = record[i]
			}
			mapData = append(mapData, line)
		}
	}

	return mapData, nil

}

// validateAttachment compares the name and date of the attachment to the issue summary and created date
// Splunk alerts attachments are created with the alertname+date.csv format, so we can compare this to the
// issue metadata to make sure it's the correct attachment
func validateAttachment(summary string, created jira.Time, attachment *jira.Attachment) bool {
	// Convert issue created time to time.Time
	a := time.Time(created)

	// Convert attachment create time to time.Time
	b, err := time.Parse("2006-01-02T15:04:05.999999+0000", attachment.Created)
	if err != nil {
		log.Println(err)
		return false
	}

	// Convert both to just time.Date
	a_date := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
	b_date := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)

	// If the attachment wasn't created on the same date as the issue, it's not the right one
	if !a_date.Equal(b_date) {
		return false
	}

	r := regexp.MustCompile(`^([A-Za-z_]+)-([0-9]{4}-[0-9]{2}-[0-9]{2}).csv$`)
	result := r.FindAllStringSubmatch(attachment.Filename, -1)

	alert_name := result[0][1]
	alert_date, err := time.Parse("2006-01-02", result[0][2])
	if err != nil {
		log.Println(err)
		return false
	}

	// If the date in the filename doesn't match the date of the issue, it's not the right one
	if !alert_date.Equal(a_date) {
		return false
	}

	// Check that the parsed attachment filename matches the summary
	r2 := regexp.MustCompile(`^Compliance Alert: ([A-Za-z\s]+)$`)
	result2 := r2.FindAllStringSubmatch(summary, -1)

	// Replace spaces with underscores in the issue summary
	r3 := regexp.MustCompile(`\s+`)
	summary_name := r3.ReplaceAllString(result2[0][1], "_")

	// If the summary name with underscores doesn't match the alert name, it's not the right one
	if summary_name != alert_name {
		return false
	}

	return true
}

// issueHasLabel returns true if the label has been applied to the issue already, else false
func issueHasLabel(issue *jira.Issue, label string) bool {
	for i := range issue.Fields.Labels {
		if issue.Fields.Labels[i] == label {
			return true
		}
	}

	return false
}
