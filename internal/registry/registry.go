package registry

import (
	"net/http"
	"time"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// VersionChecker retrieves current and latest versions for a package ecosystem.
type VersionChecker interface {
	GetLatestVersion(pkg string) (string, error)
	GetCurrentVersion(server *model.Server) (string, error)
}
