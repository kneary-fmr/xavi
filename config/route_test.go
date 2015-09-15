package config

import (
	"encoding/json"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/kvstore"
	"testing"
)

func TestJSON2Route(t *testing.T) {
	routeDef := `
		{
			"name":"route1",
			"uriRoot":"/hello",
			"backend":"hello-backend",	
			"filters":["filter1","filter2","filter3"],
			"MsgProps":"SOAPAction:\"foo\""
		}`

	var r RouteConfig

	json.Unmarshal([]byte(routeDef), &r)
	testVerifyRouteRead(&r, t)

}

func testVerifyRouteRead(r *RouteConfig, t *testing.T) {
	assert.Equal(t, "route1", r.Name)
	assert.Equal(t, "/hello", r.URIRoot)
	assert.Equal(t, "hello-backend", r.Backend)
	assert.Equal(t, 3, len(r.Filters))
	assert.Equal(t, "filter1", r.Filters[0])
	assert.Equal(t, "filter2", r.Filters[1])
	assert.Equal(t, "filter3", r.Filters[2])
	assert.Equal(t, "SOAPAction:\"foo\"", r.MsgProps)
}

func TestRouteStoreAndRetrieve(t *testing.T) {
	var testKVS, _ = kvstore.NewHashKVStore("")

	//Read - not found
	r, err := ReadRouteConfig("route1", testKVS)
	assert.Nil(t, err)
	assert.Nil(t, r, "Expected route to be nil")

	//Read - empty list
	routes, err := ListRouteConfigs(testKVS)
	assert.Nil(t, err)
	assert.Nil(t, routes)

	//Store
	var filters = []string{"filter1", "filter2", "filter3"}
	r = &RouteConfig{"route1", "/hello", "hello-backend", filters, "SOAPAction:\"foo\""}
	err = r.Store(testKVS)
	assert.Nil(t, err)

	//Read - found
	r, err = ReadRouteConfig("route1", testKVS)
	assert.Nil(t, err)
	testVerifyRouteRead(r, t)

	//Read - list
	routes, err = ListRouteConfigs(testKVS)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(routes))
	testVerifyRouteRead(routes[0], t)
}