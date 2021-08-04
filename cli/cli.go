package cli

import (
	"github.com/portainer/authenticator"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// ParseOptions parses the arguments/flags passed to the binary.
func ParseOptions() *authenticator.Options {

	options := &authenticator.Options{
		ConfigFilePath: kingpin.Flag("config", "Path to the configuration file to update").Default(authenticator.DefaultConfigFilePath).Short('c').String(),
		AddContexts:    kingpin.Flag("context", "Update / Add Portainer endpoints as Docker contexts").Bool(),
		PortainerAPI:   kingpin.Arg("portainer API URL", "URL of the Portainer API.").Required().String(),
		Username:       kingpin.Arg("Username", "Username").Required().String(),
		Password:       kingpin.Arg("Password", "Password").String(),
	}

	kingpin.Parse()

	return options
}
