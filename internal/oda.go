package internal

import (
	"embed"
)

type QueryDef struct {
	Model    string
	Filter   string
	Offset   int
	Limit    int
	Fields   string
	Count    bool
	Username string
	Password string
}

type ODA struct {
	Name    string
	Usage   string
	Version string
	EmbedFS embed.FS
	Q       QueryDef
}

func NewODA(name, usage, version string, embedFS embed.FS) *ODA {
	return &ODA{
		Name:    name,
		Usage:   usage,
		Version: version,
		EmbedFS: embedFS,
		Q:       QueryDef{},
	}
}
