package app_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi"
	"github.com/mikydna/ak/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	config := &app.Config{
		Dist: "/public",
	}

	test := chi.NewMux()
	dist := http.Dir(os.TempDir())
	test.Mount("/", app.App(config, app.ModeDev, "test", dist))

	server := httptest.NewServer(test)
	defer server.Close()

	res, err := http.Get(server.URL + "/version")
	require.NoError(t, err)

	b, err := io.ReadAll(res.Body)
	res.Body.Close()

	require.NoError(t, err)
	assert.Equal(t, string(b), "[dev, test]")
}

func TestApp_Dist(t *testing.T) {
	dist := http.FS(fstest.MapFS{
		"foo.js":  {},
		"foo.css": {},
	})

	config := &app.Config{
		Dist: "/public",
	}

	test := chi.NewMux()
	test.Mount("/", app.App(config, app.ModeDev, "test", dist))
	server := httptest.NewServer(test)
	defer server.Close()

	{
		res, err := http.Get(server.URL + "/public/foo.js")
		res.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)
	}
	{
		res, err := http.Get(server.URL + "/public/foo.js")
		res.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)
	}
	{
		res, err := http.Get(server.URL + "/public/does-not-exist")
		res.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, http.StatusNotFound)
	}
}

func TestApp_Page(t *testing.T) {
	config := &app.Config{
		Dist: "/custom-public-dir",
		Pages: []*app.PageConfig{
			{
				Route:  "/foo",
				Script: "foo.js",
				Style:  "foo.css",
				Vendor: "vendor.js",
			},
			{
				Route: "/bar",
				Style: "bar.css",
			},
			{
				Route: "/baz",
				Preload: []*app.Preload{
					{
						Href: "preload.something",
						Type: "something-type",
						As:   "something",
						Flag: "myFlag",
					},
				},
			},
		},
	}

	test := chi.NewMux()
	dist := http.Dir(os.TempDir())
	test.Mount("/", app.App(config, app.ModeDev, "test", dist))

	server := httptest.NewServer(test)
	defer server.Close()

	{
		res, err := http.Get(server.URL + "/foo")
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		b, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)

		html := string(b)

		vars := `<script>const AK = { mode: "dev", version: "test" };</script>`
		assert.True(t, strings.Contains(html, vars))

		css := `<link rel="stylesheet" href="/custom-public-dir/foo.css" />`
		assert.True(t, strings.Contains(html, css))

		js := `<script src="/custom-public-dir/foo.js" defer></script>`
		assert.True(t, strings.Contains(html, js))

		vendor := `<script src="/custom-public-dir/vendor.js" defer></script>`
		assert.True(t, strings.Contains(html, vendor))
	}

	{
		res, err := http.Get(server.URL + "/bar")
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		b, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)

		html := string(b)

		css := `<link rel="stylesheet" href="/custom-public-dir/bar.css" />`
		assert.True(t, strings.Contains(html, css))

		js := `<script src="/custom-public-dir/bar.js" defer></script>`
		assert.False(t, strings.Contains(html, js))

		vendor := `<script src="/custom-public-dir/vendor.js" defer></script>`
		assert.False(t, strings.Contains(html, vendor))
	}

	{
		res, err := http.Get(server.URL + "/baz")
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		b, err := io.ReadAll(res.Body)
		res.Body.Close()
		require.NoError(t, err)

		html := string(b)

		preload := `<link rel="preload" href="/custom-public-dir/preload.something" as="something" type="something-type" myFlag />`
		assert.True(t, strings.Contains(html, preload))
	}
}
