package app

import (
	"strings"

	"gopkg.in/yaml.v3"
)

//go:generate stringer -type=Mode -linecomment=true -output mode_str.go
type Mode uint

const (
	ModeDev  Mode = iota + 1 // dev
	ModeProd                 // prod
)

func (mode *Mode) UnmarshalYAML(node *yaml.Node) error {
	var str string
	if err := node.Decode(&str); err != nil {
		return err
	}

	*mode = ParseMode(str)

	return nil
}

func ParseMode(str string) Mode {
	var mode Mode
	switch strings.ToLower(str) {
	case "prod", "production":
		mode = ModeProd
	case "dev", "development":
		mode = ModeDev
	default:
		panic("unknown mode")
	}

	return mode
}
