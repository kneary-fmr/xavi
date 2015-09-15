package service

import (
	"bytes"
	"errors"
	"fmt"

	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
)

type route struct {
	Name             string
	URIRoot          string
	Backend          *backend
	WrapperFactories []plugin.WrapperFactory
	MsgProps         string
}

func makeRouteNotFoundError(name string) error {
	return errors.New("Route '" + name + "' not found")
}

func buildRoute(name string, kvs kvstore.KVStore) (*route, error) {
	var r route

	r.Name = name

	routeConfig, err := config.ReadRouteConfig(name, kvs)
	if err != nil {
		return nil, err
	}

	if routeConfig == nil {
		return nil, makeRouteNotFoundError(name)
	}

	backend, err := buildBackend(routeConfig.Backend, kvs)
	if err != nil {
		return nil, err
	}

	r.URIRoot = routeConfig.URIRoot
	r.Backend = backend

	for _, filterName := range routeConfig.Filters {
		factory, err := plugin.LookupWrapperFactory(filterName)
		if err != nil {
			return nil, fmt.Errorf("No wrapper factory with name %s in registry", filterName)
		}

		log.Debug("adding wrapper factory to factories")
		r.WrapperFactories = append(r.WrapperFactories, factory)

	}

	r.MsgProps = routeConfig.MsgProps

	return &r, nil
}

func (r route) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Route: %s\n", r.Name))
	buffer.WriteString(fmt.Sprintf("\tUri root: %s\n", r.URIRoot))
	buffer.WriteString(fmt.Sprintf("\tBackend: %s\n", r.Backend))
	return buffer.String()
}