package docs

import "embed"

// Docs are Intento docs.
//
//go:embed *.md */*.md
var Docs embed.FS
