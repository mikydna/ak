package app_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/mikydna/ak/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageConfig_Reader(t *testing.T) {
	config := &app.PageConfig{
		Script: "/foo.js",
	}

	ctx := context.TODO()
	w, err := config.Reader(ctx, "/custom-dist")
	require.NoError(t, err)

	b, err := io.ReadAll(w)
	require.NoError(t, err)

	html := string(b)
	script := `<script src="/custom-dist/foo.js" defer></script>`
	assert.True(t, strings.Contains(html, script))
}
