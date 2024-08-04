package tpot

import (
	"context"
	"net/http"
	"net/url"
)

type ContextBuilder[C Context] func(http.ResponseWriter, *http.Request) C

type Servable[C Context] interface {
	Handler(ContextBuilder[C]) http.Handler
	Serve(C)
}

type Middleware[C Context] func(ctx C, next func(C)) func(C)

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
