package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Config interface {
	comfig.Logger
	GetInetListener() Listener
	GetVsockListener() Listener

	GetSigner() *Signer
}

type config struct {
	comfig.Logger

	signerConfigurator comfig.Once
	inetConfigurator   comfig.Once
	vsockConfigurator  comfig.Once

	getter kv.Getter
}

func New(getter kv.Getter) Config {
	return &config{
		getter: getter,
		Logger: comfig.NewLogger(getter, comfig.LoggerOpts{}),
	}
}
