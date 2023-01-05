# Setting up a development environment

The compliance-audit-router interacts with three external services (at the moment): Splunk, LDAP and Jira. This document describes how to setup a local development environment to simulate or interact with these services.

## Splunk

The Splunk environment integration consists of two pieces:  receipt of a webook, and retrieving a search result from the webook.

Initial development was done with a CURL to the CAR running locally to simulate the webhook, and a contaier running an HTTP server that responsed with a mock Splunk search result.

TODO: There needs to be a better way to test these integrations.

### Setup Splunk config for CAR

This is currently unused for development.  The Splunk config should look something like this in production:

```yaml
splunkconfig:
  host: https://your_splunk_server:port
  username: your_splunk_username
  password: your_splunk_password
  allowinsecure: false
```

## LDAP

All of the interaction with LDAP performed by CAR are read-only operations, and very low volume, so it is probably reasonable to use the production LDAP instance for the lookups.  For Red Hat team members, this just requires you to be on the corporate VPN.

This integration is so minor that it should be possible to just mock the responses for development at some point.  As of right now, it is designed to retrieve the manager information for the user who has triggered the compliance alert.

### Setup LDAP config for CAR

Add your LDAP server connection information and search parameters to your `~/.config/compliance-audit-router/compliance-audit-router.yaml` file:

```yaml
ldapconfig:
  host: ldaps://your_ldap_server
  searchbase: dc=example,dc=com
  scope: sub
  attributes:
    - manager
```

## Jira

Atlassian offers free developer instances for testing integration with Jira and other services. [Sign up for a free development instance](http://go.atlassian.com/about-cloud-dev-instance) to test API usage.

Once your test instances has been created, log in to the instance, and navigate to [https://id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens) to create an API token.

You can validate your API token with the following:

```shell
curl -D- \
   -u your_email@example.org:your_api_token \
   -X GET \
   -H "Content-Type: application/json" \
   https://your.instance.url/rest/api/3/project
```

You should receive a 200 response with an empty project result, as nothing exists in your Jira instance yet.

### Create an Project in your instance

First retrieve your account id (note the `emailAddress` filter in the JQ command below):

```shell
curl -s \
   -u your_email@example.org:your_api_token \
   -X GET \
   -H "Content-Type: application/json" \
   https://your.instance.url/rest/api/3/users/search |jq -r '.[]|select(.emailAddress == "your_email@example.org") | .accountId'
```

Then use the accountID to create a dummy project (called OHSS in this example):

```shell
curl -u your_email@example.org:your_api_token \ 
   -X POST \
   -H "Content-Type: application/json" \
   https://your.instance.url/rest/api/3/project -d '{"key": "OHSS", "name": "OHSS (TEST)", "projectTypeKey": "software", "projectTemplateKey": "com.pyxis.greenhopper.jira:gh-kanban-template", "description": "Development OHSS Board for CAR", "leadAccountId": "your_user_account_id"}'
   ```

   This will leave you with a project that you can then tweak to configure to match your production instance.

### Setup Jira config for CAR

Add your test instance credentials to your `~/compliance-audit-router/compliance-audit-router.yaml` file:

```yaml
---
jiraconfig:
  host: https://your.instance.url
  username: your_email@example.org
  token: your_api_token
```
