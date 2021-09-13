package app

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

func distribution(mode Mode, path string, dist http.FileSystem) *chi.Mux {
	mux := chi.NewMux()
	switch mode {
	case ModeProd:
		mux.Use(productionCache(24*time.Hour, 15*time.Minute))
		mux.Use(productionBrotoli("js", "css"))
	case ModeDev:
	}

	fs := http.StripPrefix(path, http.FileServer(dist))
	mux.Handle("/*", fs)

	return mux
}

func productionCache(cacheControlMaxAge, cacheTTL time.Duration) func(next http.Handler) http.Handler {
	now := time.Now().UTC()
	cacheLastModfied := now.Format(http.TimeFormat)
	cacheExpires := now.Add(cacheTTL).Format(http.TimeFormat)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", fmt.Sprintf("max-age:%d, public", cacheControlMaxAge))
			w.Header().Set("Last-Modified", cacheLastModfied)
			w.Header().Set("Expires", cacheExpires)
			next.ServeHTTP(w, r)
		})
	}
}

func productionBrotoli(exts ...string) func(next http.Handler) http.Handler {
	// only apply to files with specified exts, ex ".(js|css)$"
	apply := regexp.MustCompile(fmt.Sprintf(".(%s)$", strings.Join(exts, "|")))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !apply.MatchString(r.URL.Path) {
				next.ServeHTTP(w, r)
			}

			w.Header().Set("Content-Encoding", "br")

			next.ServeHTTP(w, r)
		})
	}
}
