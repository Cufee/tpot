package tpot

import (
	"context"
	"net/http"
)

type ContextBuilder[C Context] func(http.ResponseWriter, *http.Request) C

type Servable[C Context] interface {
	Handler(ContextBuilder[C]) http.Handler
	Serve(C) error
}

type Middleware[C Context] func(ctx C, next func(C) error) func(C) error

func ChainHandler[C Context](ctxBuilder ContextBuilder[C], s Servable[C], middleware ...Middleware[C]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ctxBuilder(w, r)
		chain := s.Serve
		for i := len(middleware) - 1; i >= 0; i-- {
			chain = middleware[i](ctx, chain)
		}
		chain(ctx)
	})
}

type Context interface {
	Context() context.Context

	Writer() http.ResponseWriter
	Request() *http.Request
}
