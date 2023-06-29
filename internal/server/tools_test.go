package server_test

import (
	"log"
	"os"
	"testing"

	"go.infratographer.com/x/testing/eventtools"
)

var (
	nats *eventtools.TestNats
)

func setup() {
	var err error

	nats, err = eventtools.NewNatsServer()
	if err != nil {
		errPanic("failed to start nats server", err)
	}
}

func TestMain(m *testing.M) {
	setup()
	// run the tests
	code := m.Run()
	// return the test response code
	os.Exit(code)
}

func errPanic(msg string, err error) {
	if err != nil {
		log.Panicf("%s err: %s", msg, err.Error())
	}
}
