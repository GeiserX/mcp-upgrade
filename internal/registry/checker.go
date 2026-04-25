package registry

import (
	"fmt"
	"strings"
	"sync"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

func normalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

const maxConcurrent = 5

var (
	npmChecker    = NewNPMChecker()
	pypiChecker   = NewPyPIChecker()
	githubChecker = NewGitHubChecker()
	dockerChecker = NewDockerChecker()
)

func checkerForType(t model.ServerType) (VersionChecker, string) {
	switch t {
	case model.TypeNPX, model.TypeLocalNode:
		return npmChecker, ""
	case model.TypePipx, model.TypeUVX:
		return pypiChecker, ""
	case model.TypeGitHubRelease, model.TypeGoBinary:
		return githubChecker, ""
	case model.TypeDocker:
		return dockerChecker, ""
	default:
		return nil, ""
	}
}

func packageName(server *model.Server) string {
	switch server.Type {
	case model.TypeGitHubRelease, model.TypeGoBinary:
		return server.GitHubRepo
	case model.TypeDocker:
		return server.DockerImage
	default:
		return server.Package
	}
}

// CheckVersion populates version info on a single server.
func CheckVersion(server *model.Server) error {
	if server.Type == model.TypeLocal || server.Type == model.TypeUnknown {
		server.Status = model.StatusSkipped
		return nil
	}

	checker, _ := checkerForType(server.Type)
	if checker == nil {
		server.Status = model.StatusSkipped
		return nil
	}

	pkg := packageName(server)
	if pkg == "" {
		server.Status = model.StatusUnknown
		return fmt.Errorf("no package identifier for server %s", server.Name)
	}

	current, currentErr := checker.GetCurrentVersion(server)
	latest, latestErr := checker.GetLatestVersion(pkg)

	server.CurrentVersion = current
	server.LatestVersion = latest

	if currentErr != nil && latestErr != nil {
		server.Status = model.StatusError
		return fmt.Errorf("version check failed for %s: current: %v, latest: %v", server.Name, currentErr, latestErr)
	}

	if latestErr != nil {
		server.Status = model.StatusUnknown
		return nil
	}

	if current == "(auto)" || current == "(local)" {
		server.Status = model.StatusAutoLatest
		return nil
	}

	if currentErr != nil {
		server.Status = model.StatusUnknown
		return nil
	}

	if normalizeVersion(current) == normalizeVersion(latest) {
		server.Status = model.StatusUpToDate
	} else {
		server.Status = model.StatusUpgradable
	}

	return nil
}

// CheckAll checks all servers concurrently with bounded parallelism.
func CheckAll(servers []model.Server) error {
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for i := range servers {
		wg.Add(1)
		sem <- struct{}{}
		go func(s *model.Server) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := CheckVersion(s); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(&servers[i])
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("%d version check(s) failed", len(errs))
	}
	return nil
}
