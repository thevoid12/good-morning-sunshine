package assests

import "embed"

var (
	//go:embed css/*.css js/*.js  img/*.jpg img/*.svg
	AssestFS embed.FS
)
