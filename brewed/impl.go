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
type Layout[C tpot.Context] func(C, ...templ.Component) (templ.Component, error)

/*
A handler that returns a layout wrapper, and a templ.Component body
  - The layout is rendered and then wrapped around the body component
  - Both layout and component can be safely returned as nil
*/
type Page[C tpot.Context] func(C) (Layout[C], templ.Component, error)

/*
A handler that returns a body templ.Component without a layout
  - The intended use case is returning templ components for HTMX requests
  - Component can be safely returned as nil
*/
type Partial[C tpot.Context] func(C) (templ.Component, error)

/*
An endpoint handler that does not return any templ components
*/
type Endpoint[C tpot.Context] func(C) error

/*
A WebSocket specific handler
  - Returns an upgrader and a handler function that will be called after the upgrade
*/
type WebSocket[C tpot.Context] func(C) (*websocket.Upgrader, func(conn *websocket.Conn) error, error)

func (page Page[C]) Handler(ctx tpot.ContextBuilder[C]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := page.Serve(ctx(w, r)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (page Page[C]) Serve(ctx C) error {
	layout, body, err := page(ctx)
	if err != nil {
		return errors.Wrap(err, "page handler returned an error")
	}
	if layout == nil && body == nil {
		return nil
	} else if layout == nil {
		err = body.Render(ctx.Context(), ctx.Writer())
		if err != nil {
			return errors.Wrap(err, "failed to render body component")
		}
		return nil
	}

	withLayout, err := layout(ctx, body)
	if err != nil {
		return errors.Wrap(err, "layout handler returned an error")
	}
	if withLayout == nil {
		return nil
	}

	err = withLayout.Render(ctx.Context(), ctx.Writer())
	if err != nil {
		return errors.Wrap(err, "failed to render layout component")
	}

	return nil
}

func (partial Partial[C]) Handler(ctx tpot.ContextBuilder[C]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := partial.Serve(ctx(w, r)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (partial Partial[C]) Serve(ctx C) error {
	content, err := partial(ctx)
	if err != nil {
		return errors.Wrap(err, "partial handler returned an error")
	}
	if content == nil {
		return nil
	}

	err = content.Render(ctx.Context(), ctx.Writer())
	if err != nil {
		return errors.Wrap(err, "failed to render body component")
	}

	return nil
}

func (endpoint Endpoint[C]) Handler(ctx tpot.ContextBuilder[C]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := endpoint.Serve(ctx(w, r)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (endpoint Endpoint[C]) Serve(ctx C) error {
	err := endpoint(ctx)
	if err != nil {
		return errors.Wrap(err, "endpoint handler returned an error")
	}

	return nil
}

func (ws WebSocket[C]) Handler(ctx tpot.ContextBuilder[C]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := ws.Serve(ctx(w, r)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (ws WebSocket[C]) Serve(ctx C) error {
	u, handler, err := ws(ctx)
	if err != nil {
		return errors.Wrap(err, "websocket handler returned an error")
	}
	if u == nil || handler == nil {
		return nil
	}

	conn, err := u.Upgrade(ctx.Writer(), ctx.Request(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade a websocket")
	}

	return handler(conn)
}

func Redirect[C tpot.Context](url string, code int) Endpoint[C] {
	return func(ctx C) error {
		http.Redirect(ctx.Writer(), ctx.Request(), url, code)
		return nil
	}
}

func HTTP[C tpot.Context](handler http.Handler) Endpoint[C] {
	return func(ctx C) error {
		handler.ServeHTTP(ctx.Writer(), ctx.Request())
		return nil
	}
}
