package server

import "errors"

var (
	errSubscriberCreate   = errors.New("unable to create subscriber")
	errSubscriptionCreate = errors.New("unable to create subscription")
)
