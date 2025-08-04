package config

import (
	"fmt"
	"net"

	figure "gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

type inetListener struct {
	net.Listener

	Addr     string `fig:"addr,required"`
	Disabled bool   `fig:"disabled"`
}

func (l *inetListener) IsDisabled() bool {
	return l.Disabled
}

func (c *config) GetInetListener() Listener {
	return c.inetConfigurator.Do(func() any {
		var inetListener inetListener

		err := figure.
			Out(&inetListener).
			From(kv.MustGetStringMap(c.getter, "inet_listener")).
			Please()

		if err != nil {
			panic(fmt.Errorf("failed to figure out: %w", err))
		}

		if inetListener.IsDisabled() {
			return &inetListener
		}

		listener, err := net.Listen("tcp", inetListener.Addr)
		if err != nil {
			panic(fmt.Errorf("failed to listen inet on %s with error: %w", inetListener.Addr, err))
		}

		inetListener.Listener = listener

		return &inetListener
	}).(Listener)
}
