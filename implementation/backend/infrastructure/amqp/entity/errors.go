package entity

import "errors"

var (
	NotConnectedError           = errors.New("not connected to a server")
	AlreadyClosedError          = errors.New("already closed: not connected to the server")
	ShutdownError               = errors.New("client is shutting down")
	PushFailedNotConnectedError = errors.New("failed to push: not connected")
)
