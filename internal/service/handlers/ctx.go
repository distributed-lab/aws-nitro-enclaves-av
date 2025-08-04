package handlers

import (
	"context"
	"net/http"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/config"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	signerCtxKey
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxSigner(signer *config.Signer) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, signerCtxKey, signer)
	}
}

func Signer(r *http.Request) *config.Signer {
	return r.Context().Value(signerCtxKey).(*config.Signer)
}
