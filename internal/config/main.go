package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Config interface {
	comfig.Logger
	InetListenerer
	VsockListenerer

	GetSigner() *Signer
}

type config struct {
	comfig.Logger
	InetListenerer
	VsockListenerer

	signerConfigurator comfig.Once

	getter kv.Getter
}

func New(getter kv.Getter) Config {
	return &config{
		getter:          getter,
		InetListenerer:  NewInetListenerer(getter),
		VsockListenerer: NewVsockListenerer(getter),
		Logger:          comfig.NewLogger(getter, comfig.LoggerOpts{}),
	}
}
