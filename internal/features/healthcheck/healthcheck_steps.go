package healthcheck

import (
	. "github.com/lsegal/gucumber"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/internal/testsupport"
	"time"
)

func init() {

	//Endpoints associated with the test
	var (
		xaviAgentURL = testsupport.XaviAgentRESTEnpointBaseURI
		testUrl      = testsupport.XaviAcceptanceTestEndpointBaseURL + "/hello"
		server2Url   = testsupport.CohostedMountebankEndpointBaseURL + "/imposters/3100"
	)

	//XAVI definitions for the test scenario
	const (
		hello1Server = `{"Address":"localhost","Port":3000,"PingURI":"/hello","HealthCheck":"http-get",
							"HealthCheckInterval":200,"HealthCheckTimeout":150}`
		hello2Server = `{"Address":"localhost","Port":3100,"PingURI":"/hello","HealthCheck":"http-get",
							"HealthCheckInterval":200,"HealthCheckTimeout":150}`
		backend  = `{"ServerNames":["hello1","hello2"],"LoadBalancerPolicy":"round-robin"}`
		route    = `{"URIRoot":"/hello","Backend":"demo-backend","Filters":null,"MsgProps":""}`
		listener = `{"RouteNames":["demo-route"]}`
	)

	var (
		failedState bool
		failures    int
		spawnedPID  int
	)

	var doSetup = func() error {
		testPort, err := testsupport.GetPortFromURL(testUrl)
		if err != nil {
			return err
		}

		log.Info("Delete mb imposter")
		testsupport.DeleteMountebankImposter(server2Url)

		//
		// XAVI configuration for the test
		//
		err = testsupport.PutDefinitionOk("v1/servers/hello1", hello1Server, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/servers/hello2", hello2Server, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/backends/demo-backend", backend, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/routes/demo-route", route, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/listeners/demo-listener", listener, xaviAgentURL)
		if err != nil {
			return err
		}

		//
		// Configuration of the mountebank imposters that represent the servers proxied in
		// the test. Note we only set up one imposter to simulate the unhealhy endpoint
		//
		log.Info("Set up the healthy server on port 3000")
		testsupport.MountebankSetup(testsupport.CohostedMountebankEndpointBaseURL+"/imposters", testsupport.RoundRobin3000Config)

		spawnedPID, err = testsupport.Spawn("demo-listener", testPort, xaviAgentURL)
		log.Info("spawnedPID is ", spawnedPID)
		return err
	}

	Given(`^A backend with some unhealthy servers$`, func() {
		if err := doSetup(); err != nil {
			log.Info("Setup failed: ", err.Error())
			T.Errorf("Error in test setup: %s", err.Error())
			failedState = true
			return
		}
	})

	And(`^I invoke a service against the backend$`, func() {
		if failedState {
			return
		}
		time.Sleep(2 * time.Second)
		for i := 0; i < 5; i++ {
			log.Println("send request")
			if testsupport.GetTestEndpoint(testUrl) != 200 {
				failures = failures + 1
			}
		}
	})

	Then(`^The service calls succeed against the healthy backends$`, func() {
		if failedState {
			return
		}
		assert.Equal(T, 0, failures)
	})

	Given(`^A previously unhealthy server becomes healthy$`, func() {
		if failedState {
			T.Errorf("requisite test set up failed")
			return
		}
		log.Info("Set up a healthy server on port 3100")
		testsupport.MountebankSetup(testsupport.CohostedMountebankEndpointBaseURL+"/imposters", testsupport.RoundRobin3100Config)
	})

	Then(`^The healed backend recieves traffic$`, func() {
		if failedState {
			return
		}
		failures = 0
		time.Sleep(2 * time.Second)
		for i := 0; i < 5; i++ {
			log.Info("get ", testUrl)
			if testsupport.GetTestEndpoint(testUrl) != 200 {
				failures = failures + 1
			}
			assert.Equal(T, 0, failures)
		}

	})

	After("@withhealed", func() {
		testPort, err := testsupport.GetPortFromURL(testUrl)
		assert.NotNil(T, err)
		testsupport.KillSpawnedProcess(spawnedPID, testPort, xaviAgentURL)
	})

}