package brewed

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/cufee/tpot"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

/*
A convenience wrapper for rendering a layout with some customizations
*/
type Layout func(tpot.Context, ...templ.Component) (templ.Component, error)

var _ tpot.Servable = new(Page)

/*
A handler that returns a layout wrapper, and a templ.Component body
  - The layout is rendered and then wrapped around the body component
  - Both layout and component can be safely returned as nil
*/
type Page func(tpot.Context) (Layout, templ.Component, error)

var _ tpot.Servable = new(Partial)

/*
A handler that returns a body templ.Component without a layout
  - The intended use case is returning templ components for HTMX requests
  - Component can be safely returned as nil
*/
type Partial func(tpot.Context) (templ.Component, error)

var _ tpot.Servable = new(Endpoint)

/*
An endpoint handler that does not return any templ components
*/
type Endpoint func(tpot.Context) error

var _ tpot.Servable = new(WebSocket)

/*
A WebSocket specific handler
  - Returns an upgrader and a handler function that will be called after the upgrade
*/
type WebSocket func(tpot.Context) (*websocket.Upgrader, func(conn *websocket.Conn) error, error)

func (page Page) Handler(ctx tpot.ContextBuilder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page.Serve(ctx(w, r))
	})
}

func (page Page) Serve(ctx tpot.Context) {
	layout, body, err := page(ctx)
	if err != nil {
		ctx.Err(errors.Wrap(err, "page handler returned an error"))
		return
	}
	if layout == nil && body == nil {
		return
	} else if layout == nil {
		err = body.Render(ctx.Ctx(), ctx.Writer())
		if err != nil {
			ctx.Err(errors.Wrap(err, "failed to render body component"))
			return
		}
		return
	}

	withLayout, err := layout(ctx, body)
	if err != nil {
		ctx.Err(errors.Wrap(err, "layout handler returned an error"))
		return
	}
	if withLayout == nil {
		return
	}

	err = withLayout.Render(ctx.Ctx(), ctx.Writer())
	if err != nil {
		ctx.Err(errors.Wrap(err, "failed to render layout component"))
		return
	}
}

func (partial Partial) Handler(ctx tpot.ContextBuilder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		partial.Serve(ctx(w, r))
	})
}

func (partial Partial) Serve(ctx tpot.Context) {
	content, err := partial(ctx)
	if err != nil {
		ctx.Err(errors.Wrap(err, "partial handler returned an error"))
		return
	}
	if content == nil {
		return
	}

	err = content.Render(ctx.Ctx(), ctx.Writer())
	if err != nil {
		ctx.Err(errors.Wrap(err, "failed to render body component"))
		return
	}
}

func (endpoint Endpoint) Handler(ctx tpot.ContextBuilder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint.Serve(ctx(w, r))
	})
}

func (endpoint Endpoint) Serve(ctx tpot.Context) {
	err := endpoint(ctx)
	if err != nil {
		ctx.Err(errors.Wrap(err, "endpoint handler returned an error"))
		return
	}
}

func (ws WebSocket) Handler(ctx tpot.ContextBuilder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.Serve(ctx(w, r))
	})
}

func (ws WebSocket) Serve(ctx tpot.Context) {
	u, handler, err := ws(ctx)
	if err != nil {
		ctx.Err(errors.Wrap(err, "websocket handler returned an error"))
		return
	}
	if u == nil || handler == nil {
		return
	}

	conn, err := u.Upgrade(ctx.Writer(), ctx.Request(), nil)
	if err != nil {
		ctx.Err(errors.Wrap(err, "failed to upgrade a websocket"))
		return
	}
	handler(conn)
}

func Redirect(url string, code int) Endpoint {
	return func(ctx tpot.Context) error {
		ctx.Redirect(url, code)
		return nil
	}
}

func HTTP(handler http.Handler) Endpoint {
	return func(ctx tpot.Context) error {
		handler.ServeHTTP(ctx.Writer(), ctx.Request())
		return nil
	}
}
