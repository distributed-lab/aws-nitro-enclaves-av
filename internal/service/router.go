package service

import (
	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/service/handlers"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
)

func (s *service) router() chi.Router {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(s.log),
		ape.LoganMiddleware(s.log),
		ape.CtxMiddleware(
			handlers.CtxLog(s.log),
			handlers.CtxSigner(s.signer),
		),
	)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/attestations", handlers.VerifyAttestation)
		r.Route("/attestation-documents", func(r chi.Router) {
			r.Get("/address", handlers.GetAddressAttestationDoc)
			r.Get("/public-key", handlers.GetPublicKeyAttestationDoc)
		})
	})

	return r
}
