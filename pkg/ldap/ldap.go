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
	"log"

	"github.com/go-ldap/ldap"
	"github.com/openshift/compliance-audit-router/pkg/config"
)

type ConnectionLayer interface {
	Close()
	SimpleBind(ldapAddr string) (*ldap.Conn, error)
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

// LDAPDataAccessLayer implements the ConnectionLayer for LDAP
type LDAPDataAccessLayer struct {
	conn    *ldap.Conn
	address string
}

// NewLDAPDataAccessLayer creates a new LDAPAccessLayer
func NewLDAPDataAccessLayer(address string) (*LDAPDataAccessLayer, error) {
	conn, err := ldap.DialURL(address)
	ldapDAL := &LDAPDataAccessLayer{
		conn:    conn,
		address: address,
	}

	return ldapDAL, err
}

// Close closes the connection to the LDAP server
func (l *LDAPDataAccessLayer) Close() {
	l.conn.Close()
}

// SimpleBind performs a simple bind to the LDAP server
//func (l *LDAPDataAccessLayer) SimpleBind(ldapAddr string) (*ldap.Conn, error) {
//	return l.ldap.DialURL(ldapAddr)
//
//}

// LookupUser performs an LDAP query to find the user's supplemental ID and manager information
func LookupUser(username string) (string, string, error) {

	ldapAddr := buildLDAPAddr(config.AppConfig.LDAPConfig.Host, config.AppConfig.LDAPConfig.Port)

	conn, err := ldap.DialURL(ldapAddr)
	defer conn.Close()
	if err != nil {
		return "", "", err
	}

	if config.AppConfig.LDAPConfig.Username != "" {
		_, err = conn.SimpleBind(&ldap.SimpleBindRequest{
			Username: config.AppConfig.LDAPConfig.Username,
			Password: config.AppConfig.LDAPConfig.Password,
		})
		fmt.Println("DEBUG 4a")
	} else {
		err = conn.UnauthenticatedBind("")
		fmt.Println("DEBUG 4b")
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
	fmt.Println("DEBUG 6")

	if len(result.Entries) == 0 {
		return "", "", errors.New("User not found")
	} else {
		for _, entry := range result.Entries {
			for _, attribute := range entry.Attributes {
				if len(attribute.Values) > 1 {
					log.Printf("multiple attributes found for user %s: %s - using the first (%s)",
						username, attribute.Name, attribute.Values)
				}
			}
		}
	}

	return "", "", errors.New("Not implemented")
}

// buildLDAPAddr creates the LDAP URL from the host and port provided in the config
func buildLDAPAddr(host string, port int) string {
	if port != 0 {
		return fmt.Sprintf("%s:%d", host, port)
	}
	return host
}
