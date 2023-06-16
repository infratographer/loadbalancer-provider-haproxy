//go:build testtools
// +build testtools

package ipam_test

import (
	"log"
	"os"
	"testing"

	"go.infratographer.com/x/events"
	"go.infratographer.com/x/testing/eventtools"
)

var SC events.SubscriberConfig

func TestMain(m *testing.M) {
	setup()
	// run the tests
	code := m.Run()
	// return the test response code
	os.Exit(code)
}

func setup() {
	var err error
	_, SC, err = eventtools.NewNatsServer()

	if err != nil {
		errPanic("failed to start nats server", err)
	}
}

func errPanic(msg string, err error) {
	if err != nil {
		log.Panicf("%s err: %s", msg, err.Error())
	}
}
