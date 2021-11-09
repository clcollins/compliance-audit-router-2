package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

// Package config loads configuration details so they can be accessed
// by other packages

var Appname = "compliance-audit-router"
var defaultMessageTemplate = "{{.Username}} and {{.Manager}}\n\n" +
	"This action requires justification." +
	"Please provide the justification in the comments section below."

var AppConfig Config

type Config struct {
	Verbose         bool
	ListenPort      int
	MessageTemplate string

	LDAPConfig   LDAPConfig
	SplunkConfig SplunkConfig
	JiraConfig   JiraConfig
}

type LDAPConfig struct {
	Host               string
	Port               int
	InsecureSkipVerify bool
	Username           string
	Password           string
	SearchBase         string
	Scope              string
	Attributes         []string
}

type SplunkConfig struct {
	Host          string
	Port          int
	AllowInsecure bool
}

type JiraConfig struct {
	Host          string
	Port          int
	AllowInsecure bool
	Username      string
	Password      string
	Project       string
}

func init() {

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	viper.AddConfigPath(home + "/.config/" + Appname) // Look for config in $HOME/.config/compliance-audit-router
	viper.SetConfigType("yaml")
	viper.SetConfigName(Appname)

	viper.SetEnvPrefix("car")
	viper.AutomaticEnv() // read in environment variables that match

	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(err)
	}

	log.Printf("Using config file: %s", viper.ConfigFileUsed())

	viper.SetDefault("MessageTemplate", defaultMessageTemplate)
	viper.SetDefault("Verbose", true)
	viper.SetDefault("ListenPort", 8080)
	viper.SetDefault("LDAPConfig.Port", 636)

	err = viper.Unmarshal(&AppConfig)
	if err != nil {
		panic(err)
	}

}

func GetLDAPAddr() string {
	return AppConfig.LDAPConfig.Host + ":" + fmt.Sprint(AppConfig.LDAPConfig.Port)
}
