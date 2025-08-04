package config

import (
	"fmt"
	"net"

	"github.com/mdlayher/vsock"
	figure "gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

type vsockListener struct {
	net.Listener

	ContextID uint32 `fig:"context_id"`
	Port      uint32 `fig:"port,required"`
	Disabled  bool   `fig:"disabled"`
}

func (l *vsockListener) IsDisabled() bool {
	return l.Disabled
}

func (c *config) GetVsockListener() Listener {
	return c.vsockConfigurator.Do(func() any {
		var vsockListener vsockListener

		err := figure.
			Out(&vsockListener).
			From(kv.MustGetStringMap(c.getter, "vsock_listener")).
			Please()

		if err != nil {
			panic(fmt.Errorf("failed to figure out: %w", err))
		}

		if vsockListener.IsDisabled() {
			return &vsockListener
		}

		listener, err := vsock.ListenContextID(vsockListener.ContextID, vsockListener.Port, nil)
		if err != nil {
			panic(fmt.Errorf("failed to listen vsock on %d:%d with error: %w", vsockListener.ContextID, vsockListener.Port, err))
		}

		vsockListener.Listener = listener

		return &vsockListener
	}).(Listener)
}
