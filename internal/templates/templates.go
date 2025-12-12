package templates

import "embed"

//go:embed index.html
var IndexHTML embed.FS

//go:embed assets/*
var AssetHTML embed.FS
