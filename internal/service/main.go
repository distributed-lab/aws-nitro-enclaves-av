package service

import (
	"net/http"
	"sync"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/config"
	"gitlab.com/distributed_lab/logan/v3"
)

type service struct {
	log    *logan.Entry
	signer *config.Signer

	inetListener  config.Listener
	vsockListener config.Listener
}

func (s *service) run() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		if s.inetListener.IsDisabled() {
			s.log.Warn("Inet listener disabled")
			wg.Done()
			return
		}

		r := s.router()

		s.log.Info("Inet listener started")
		if err := http.Serve(s.inetListener, r); err != nil {
			s.log.WithError(err).Error("Inet serve exit with error")
		}
		s.log.Info("Inet listener stopped")

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		if s.inetListener.IsDisabled() {
			s.log.Warn("Vsock listener disabled")
			wg.Done()
			return
		}

		r := s.router()

		s.log.Info("Vsock listener started")
		if err := http.Serve(s.vsockListener, r); err != nil {
			s.log.WithError(err).Error("Vsock serve exit with error")
		}
		s.log.Info("Vsock listener stopped")

		wg.Done()
	}()

	s.log.Info("Service started")
	wg.Wait()
	s.log.Info("Service stopped")
	return nil
}

func newService(cfg config.Config) *service {
	return &service{
		log:    cfg.Log(),
		signer: cfg.GetSigner(),

		inetListener:  cfg.GetInetListener(),
		vsockListener: cfg.GetVsockListener(),
	}
}

func Run(cfg config.Config) {
	if err := newService(cfg).run(); err != nil {
		panic(err)
	}
}
