package commands

import (
	"flag"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/mitchellh/cli"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
	"strings"
)

//AddRoute command
type AddRoute struct {
	UI      cli.Ui
	KVStore kvstore.KVStore
}

//Help provides details on command line options for AddRoute
func (ar *AddRoute) Help() string {
	helpText := `
	Usage: xavi [options]

	Options
		-name Route name
		-backend Backend name
		-base-uri Base uri to match
		-filters Optional list of filter names
		-msgprop Message properties for matching route
		`

	return strings.TrimSpace(helpText)
}

func (ar *AddRoute) validateBackend(name string) (bool, error) {
	key := "backends/" + name
	log.Info("Read key " + key)
	backend, err := ar.KVStore.Get(key)
	if err != nil {
		return false, err
	}

	return backend != nil, nil
}

func filtersRegistered(filters []string) (string, bool) {
	if len(filters) > 0 && filters[0] != "" {
		for _, f := range filters {
			if !plugin.RegistryContains(f) {
				return f, false
			}
		}
	}

	return "", true
}

//Run executes the AddRoute command using the provided arguments
func (ar *AddRoute) Run(args []string) int {
	var name, backend, baseuri, filterList, msgprop string
	cmdFlags := flag.NewFlagSet("add-route", flag.ContinueOnError)
	cmdFlags.Usage = func() { ar.UI.Output(ar.Help()) }
	cmdFlags.StringVar(&name, "name", "", "")
	cmdFlags.StringVar(&backend, "backend", "", "")
	cmdFlags.StringVar(&baseuri, "base-uri", "", "")
	cmdFlags.StringVar(&filterList, "filters", "", "")
	cmdFlags.StringVar(&msgprop, "msgprop", "", "")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	argErr := false

	if name == "" {
		ar.UI.Error("Name must be specified")
		argErr = true
	}

	if backend == "" {
		ar.UI.Error("Backend must be specified")
		argErr = true
	}

	if baseuri == "" {
		ar.UI.Error("Base uri must be specified")
		argErr = true
	}

	if argErr {
		ar.UI.Error("")
		ar.UI.Error(ar.Help())
		return 1
	}

	validName, err := ar.validateBackend(backend)
	if err != nil || !validName {
		ar.UI.Error("backend not found: " + name)
		return 1
	}

	var filters []string
	if filterList != "" {
		filters = strings.Split(filterList, ",")
		unregistered, filtersRegistered := filtersRegistered(filters)
		if !filtersRegistered {
			ar.UI.Error("Error: filter list contains unregistered filter: '" + unregistered + "'")
			return 1
		}
	}

	route := &config.RouteConfig{
		Name:     name,
		Backend:  backend,
		URIRoot:  baseuri,
		Filters:  filters,
		MsgProps: msgprop,
	}

	if err := route.Store(ar.KVStore); err != nil {
		ar.UI.Error(err.Error())
		return 1
	}

	if err := ar.KVStore.Flush(); err != nil {
		ar.UI.Error(err.Error())
		return 1
	}

	return 0
}

//Synopsis provides a concise description of the AddRoute command.
func (ar *AddRoute) Synopsis() string {
	return "Create a route linking a uri pattern to a backend"
}
