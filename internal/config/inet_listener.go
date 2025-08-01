package config

import (
	"fmt"
	"net"

	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type InetListenerer interface {
	InetListener() net.Listener
}

func NewInetListenerer(getter kv.Getter) InetListenerer {
	return &inetListener{
		getter: getter,
	}
}

type inetListener struct {
	once   comfig.Once
	getter kv.Getter
}

func (c *inetListener) InetListener() net.Listener {
	return c.once.Do(func() interface{} {
		var cfg struct {
			Addr     string `fig:"addr,required"`
			Disabled bool   `fig:"disabled"`
		}

		err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(c.getter, "inet_listener")).
			Please()

		if err != nil {
			panic(fmt.Errorf("failed to figure out: %w", err))
		}

		listener, err := net.Listen("tcp", cfg.Addr)
		if err != nil {
			panic(fmt.Errorf("failed to listen inet on %s with error: %w", cfg.Addr, err))
		}

		return listener
	}).(net.Listener)
}
