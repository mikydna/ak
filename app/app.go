package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type Config struct {
	Mount string
	Dist  string
	Pages []*PageConfig
}

func App(cfg *Config, mode Mode, version string, dist http.FileSystem) *chi.Mux {
	mux := chi.NewMux()

	// middleware
	mux.Use(modeAndVersion(mode, version))

	switch mode {
	case ModeProd:
		mux.Use(production)
	case ModeDev:
		mux.Use(development)
	}

	// helpers
	mux.Get("/version", modeAndVersionHandlerFunc(mode, version))

	// statics (js, css, etc...)
	if dist != nil {
		mux.Mount(cfg.Dist, distribution(mode, cfg.Dist, dist))
	}

	// end-user pages
	for _, curr := range cfg.Pages {
		if dist == nil {
			panic("pages require a distribution")
		}

		mux.With(asHTML).Get(curr.Route, page(curr, cfg.Dist))
	}

	return mux
}

// ***** middleware *****

func asHTML(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		next.ServeHTTP(w, r)
	})
}

func production(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age:900, public")
		next.ServeHTTP(w, r)
	})
}

func development(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// ***** app ctx setters and helpers *****

type CtxVersion struct{}
type CtxMode struct{}

func modeAndVersion(mode Mode, version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rCtx := context.WithValue(r.Context(), CtxVersion{}, version)
			rCtx = context.WithValue(rCtx, CtxMode{}, mode.String())
			next.ServeHTTP(w, r.WithContext(rCtx))
		})
	}
}

func modeAndVersionHandlerFunc(mode Mode, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		str := fmt.Sprintf("[%s, %s]", mode.String(), version)
		fmt.Fprint(w, str)
	}
}
