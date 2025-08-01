package config

import (
	"fmt"
	"net"

	"github.com/mdlayher/vsock"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type VsockListenerer interface {
	VsockListener() net.Listener
}

func NewVsockListenerer(getter kv.Getter) VsockListenerer {
	return &vsockListener{
		getter: getter,
	}
}

type vsockListener struct {
	once   comfig.Once
	getter kv.Getter
}

func (c *vsockListener) VsockListener() net.Listener {
	return c.once.Do(func() interface{} {
		var cfg struct {
			ContextID uint32 `fig:"context_id"`
			Port      uint32 `fig:"port,required"`
			Disabled  bool   `fig:"disabled"`
		}

		err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(c.getter, "vsock_listener")).
			Please()

		if err != nil {
			panic(fmt.Errorf("failed to figure out: %w", err))
		}

		listener, err := vsock.ListenContextID(cfg.ContextID, cfg.Port, nil)
		if err != nil {
			panic(fmt.Errorf("failed to listen vsock on %d:%d with error: %w", cfg.ContextID, cfg.Port, err))
		}

		return listener
	}).(net.Listener)
}
