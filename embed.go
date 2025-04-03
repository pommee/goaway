//go:build embed

package main

import (
	"embed"
)

//go:embed website/dist/*
var embeddedContent embed.FS

func init() {
	content = embeddedContent
}
