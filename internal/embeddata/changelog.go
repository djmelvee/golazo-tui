package embeddata

import _ "embed"

// Changelog is embedded at build time so the in-app viewer works outside the project root.
//
//go:embed CHANGELOG.md
var Changelog []byte