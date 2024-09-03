package assests

import "embed"

var (
	//go:embed css/*.css js/*.js  img/*.jpg
	AssestFS embed.FS
)
