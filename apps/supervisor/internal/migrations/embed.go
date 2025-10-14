package migrations

import "embed"

// Files exposes the supervisor SQL migrations packaged with the service.
//
//go:embed sql/*.sql
var Files embed.FS

// Dir is the root directory containing supervisor migrations within the embedded filesystem.
const Dir = "sql"
