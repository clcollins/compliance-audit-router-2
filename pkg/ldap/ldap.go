package ldap

import (
	"errors"
	"fmt"

	"github.com/go-ldap/ldap"
	"github.com/openshift/compliance-audit-router/pkg/config"
)

func LookupUser(username string) (string, string, error) {
	ldapAddr := config.GetLDAPAddr()

	fmt.Println(ldapAddr)
	fmt.Println("DEBUG 2")

	conn, err := ldap.DialURL("ldap://" + ldapAddr)
	defer conn.Close()
	if err != nil {
		return "", "", err
	}

	fmt.Println("DEBUG 3")

	if config.AppConfig.LDAPConfig.Username != "" {
		fmt.Println("DEBUG 4")
		_, err = conn.SimpleBind(&ldap.SimpleBindRequest{
			Username: config.AppConfig.LDAPConfig.Username,
			Password: config.AppConfig.LDAPConfig.Password,
		})
		fmt.Println("DEBUG 4a")
	} else {
		fmt.Println("DEBUG 5")
		conn.UnauthenticatedBind("")
		fmt.Println("DEBUG 5a")
	}
	if err != nil {
		return "", "", err
	}

	fmt.Println("DEBUG 6")
	searchRequest := ldap.NewSearchRequest(config.AppConfig.LDAPConfig.SearchBase, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, "(uid="+username+")", config.AppConfig.LDAPConfig.Attributes, nil)

	fmt.Println("DEBUG 7")
	result, err := conn.Search(searchRequest)
	if err != nil {
		return "", "", err
	}

	fmt.Println("DEBUG 8")
	fmt.Printf("%+v\n", &result)

	return "", "", errors.New("Not implemented")
}
