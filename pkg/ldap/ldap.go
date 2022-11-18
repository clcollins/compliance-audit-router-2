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

package ldap

import (
	"errors"
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/openshift/compliance-audit-router/pkg/config"
)

type ConnectionLayer interface {
	Close()
	SimpleBind(ldapAddr string) (*ldap.Conn, error)
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

// DataAccessLayer implements the ConnectionLayer for LDAP
type DataAccessLayer struct {
	conn    *ldap.Conn
	address string
}

// NewLDAPDataAccessLayer creates a new LDAPAccessLayer
func NewLDAPDataAccessLayer(address string) (*DataAccessLayer, error) {
	conn, err := ldap.DialURL(address)
	ldapDAL := &DataAccessLayer{
		conn:    conn,
		address: address,
	}

	return ldapDAL, err
}

// Close closes the connection to the LDAP server
func (l *DataAccessLayer) Close() {
	l.conn.Close()
}

// SimpleBind performs a simple bind to the LDAP server
//func (l *LDAPDataAccessLayer) SimpleBind(ldapAddr string) (*ldap.Conn, error) {
//	return l.ldap.DialURL(ldapAddr)
//
//}

// LookupUser performs an LDAP query to find the user's supplemental ID and manager information
func LookupUser(username string) (string, string, error) {

	conn, err := ldap.DialURL(config.AppConfig.LDAPConfig.Host)
	defer conn.Close()
	if err != nil {
		return "", "", err
	}

	if config.AppConfig.LDAPConfig.Username != "" {
		_, err = conn.SimpleBind(&ldap.SimpleBindRequest{
			Username: config.AppConfig.LDAPConfig.Username,
			Password: config.AppConfig.LDAPConfig.Password,
		})
	} else {
		err = conn.UnauthenticatedBind("")
	}
	if err != nil {
		return "", "", err
	}

	searchRequest := ldap.NewSearchRequest(config.AppConfig.LDAPConfig.SearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(uid="+username+")", config.AppConfig.LDAPConfig.Attributes, nil)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return "", "", err
	}

	var ldapUsername string
	var ldapManager string

	if len(result.Entries) == 0 {
		return "", "", errors.New("User not found")
	} else if len(result.Entries) > 1 {
		return "", "", errors.New("multiple ldap entries found, please check your ldap config")
	} else {
		entry := result.Entries[0]

		ldapUsername, err = getUID(entry.DN)
		if err != nil {
			return "", "", errors.New("could not parse ldap username")
		}
		ldapManager, err = getUID(entry.GetAttributeValue("manager"))
		if err != nil {
			return "", "", errors.New("could not parse manager's ldap username")
		}
	}

	return ldapUsername, ldapManager, nil
}

func getUID(dn string) (string, error) {
	parsedDN, err := ldap.ParseDN(dn)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error parsing dn: %v", err))
	}
	for _, rdn := range parsedDN.RDNs {
		for _, attribute := range rdn.Attributes {
			if attribute.Type == "uid" {
				return attribute.Value, nil
			}
		}
	}
	return "", errors.New("no uid field found for given ldap string")
}
