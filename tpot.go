package tpot

import (
	"context"
	"net/http"
	"net/url"
)

type Servable interface {
	Serve(func() Context)
}

type Middleware func(ctx func() Context, next func(func() Context)) func(func() Context)

func Handler(ctxBuilder func() Context, s Servable, middleware ...Middleware) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentCtx := ctxBuilder()
		ctx := func() Context { return parentCtx }

		chain := s.Serve
		for i := len(middleware) - 1; i >= 0; i-- {
			chain = middleware[i](ctx, chain)
		}
		chain(ctx)
	})
}

type Context interface {
	Ctx() context.Context

	Writer() http.ResponseWriter
	Request() *http.Request

	RealIP() (string, bool)

	URL() *url.URL
	PathValue(key string) string

	SetHeader(key, value string)
	GetHeader(key string) string

	Cookie(key string) (*http.Cookie, error)
	SetCookie(cookie *http.Cookie)

	Query() (url.Values, error)
	QueryValue(key string) string

	Form() (url.Values, error)
	FormValue(key string) (string, error)

	Err(err error)
	Error(format string, args ...any)

	String(format string, args ...any)

	SetStatus(code int)
	Redirect(path string, code int)
}
