package app_test

import (
	"bytes"
	"testing"

	"github.com/mikydna/ak/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMode_UnmarshalYAML(t *testing.T) {
	var config *struct {
		Foo app.Mode `yaml:"mode"`
	}

	{
		configStr := "mode: dev\n"
		r := bytes.NewBufferString(configStr)
		err := yaml.NewDecoder(r).Decode(&config)
		require.NoError(t, err)
		assert.Equal(t, config.Foo, app.ModeDev)
	}

	{
		configStr := "mode: prod\n"
		r := bytes.NewBufferString(configStr)
		err := yaml.NewDecoder(r).Decode(&config)
		require.NoError(t, err)
		assert.Equal(t, config.Foo, app.ModeProd)
	}
}
