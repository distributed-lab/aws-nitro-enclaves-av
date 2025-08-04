package config

import (
	"errors"
	"net"
)

var (
	ErrListenerDisabled       = errors.New("listener is disabled")
	ErrListenerNotInitialized = errors.New("listener is not initialized")
)

type Listener interface {
	net.Listener
	IsDisabled() bool
}
