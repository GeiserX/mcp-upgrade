package detect

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// DetectType analyzes a server's Command and Args to determine its ServerType
// and populate related fields (Package, BinaryPath, DockerImage).
func DetectType(s *model.Server) {
	cmd := filepath.Base(s.Command)

	switch {
	case cmd == "npx":
		detectNPX(s)
	case cmd == "node":
		s.Type = model.TypeLocalNode
	case cmd == "uvx":
		detectUVX(s)
	case cmd == "python" || cmd == "python3":
		detectPython(s)
	case cmd == "docker":
		detectDocker(s)
	case strings.Contains(s.Command, "/go/bin/"):
		s.Type = model.TypeGoBinary
		s.BinaryPath = resolveAbsPath(s.Command)
	case cmd == "cargo":
		s.Type = model.TypeCargo
	case isExecutableFile(s.Command):
		s.Type = model.TypeGitHubRelease
		s.BinaryPath = resolveAbsPath(s.Command)
	default:
		s.Type = model.TypeUnknown
	}
}

// DetectAll runs DetectType on each server in the slice.
func DetectAll(servers []model.Server) {
	for i := range servers {
		DetectType(&servers[i])
	}
}

func detectNPX(s *model.Server) {
	s.Type = model.TypeNPX
	// Find the package name: skip flags and the -y/--yes marker.
	skipNext := false
	for _, arg := range s.Args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "-y" || arg == "--yes" {
			continue
		}
		if strings.HasPrefix(arg, "-") {
			// Some flags consume the next arg (e.g. --package foo).
			if arg == "--package" || arg == "-p" {
				skipNext = true
			}
			continue
		}
		s.Package = stripVersionSuffix(arg)
		return
	}
}

func detectUVX(s *model.Server) {
	s.Type = model.TypeUVX
	for _, arg := range s.Args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		s.Package = stripVersionSuffix(arg)
		return
	}
}

func detectPython(s *model.Server) {
	for i, arg := range s.Args {
		if arg == "-m" && i+1 < len(s.Args) {
			s.Type = model.TypePipx
			s.Package = s.Args[i+1]
			return
		}
	}
	s.Type = model.TypeUnknown
}

func detectDocker(s *model.Server) {
	hasRun := false
	for _, arg := range s.Args {
		if arg == "run" {
			hasRun = true
			break
		}
	}
	if !hasRun {
		s.Type = model.TypeUnknown
		return
	}

	s.Type = model.TypeDocker

	// Walk args after "run" to find the image (last non-flag positional arg
	// before any command that follows the image).
	afterRun := false
	skipNext := false
	var candidate string
	for _, arg := range s.Args {
		if !afterRun {
			if arg == "run" {
				afterRun = true
			}
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--" {
			break
		}
		if strings.HasPrefix(arg, "-") {
			// Flags that consume the next argument.
			if isFlagWithValue(arg) {
				skipNext = true
			}
			continue
		}
		candidate = arg
	}
	s.DockerImage = candidate
}

// isFlagWithValue returns true for docker-run flags that take a separate value argument.
func isFlagWithValue(flag string) bool {
	// If the flag already contains '=', the value is inline.
	if strings.Contains(flag, "=") {
		return false
	}
	valuedFlags := []string{
		"-e", "--env",
		"-v", "--volume",
		"-p", "--publish",
		"-w", "--workdir",
		"--name",
		"--network",
		"--entrypoint",
		"--mount",
		"-l", "--label",
		"-u", "--user",
		"--platform",
		"--cpus",
		"-m", "--memory",
	}
	for _, f := range valuedFlags {
		if flag == f {
			return true
		}
	}
	return false
}

// stripVersionSuffix removes a trailing @version from a package name.
// E.g. "@org/pkg@latest" -> "@org/pkg", "pkg@1.2.3" -> "pkg".
// Scoped packages like "@org/pkg" (no version) are returned as-is.
func stripVersionSuffix(name string) string {
	if name == "" {
		return name
	}

	// For scoped packages (@org/pkg@version), the version suffix is after
	// the last '@' only if there are at least two '@' characters.
	if strings.HasPrefix(name, "@") {
		lastAt := strings.LastIndex(name, "@")
		if lastAt > 0 {
			return name[:lastAt]
		}
		return name
	}

	// For unscoped packages, strip anything after '@'.
	if idx := strings.Index(name, "@"); idx > 0 {
		return name[:idx]
	}
	return name
}

func resolveAbsPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0111 != 0
}

var knownBinaries = map[string]string{
	"github-mcp-server":    "github/github-mcp-server",
	"terraform-mcp-server": "hashicorp/terraform-mcp-server",
	"mcp-atlassian":        "sooperset/mcp-atlassian",
}

func resolveGitHubRepo(s *model.Server) {
	name := filepath.Base(s.Command)
	if repo, ok := knownBinaries[name]; ok {
		s.GitHubRepo = repo
	}
}

// ResolveMetadata populates GitHubRepo from known binary mappings
// and detects pipx binaries.
func ResolveMetadata(servers []model.Server) {
	for i := range servers {
		s := &servers[i]
		if s.Type == model.TypeGitHubRelease || s.Type == model.TypeGoBinary {
			resolveGitHubRepo(s)
		}
		binName := filepath.Base(s.Command)
		if s.Type == model.TypeGitHubRelease && isPipxBinary(binName) {
			s.Type = model.TypePipx
			s.Package = binName
		}
	}
}

func isPipxBinary(binaryName string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	venvPath := filepath.Join(home, ".local", "pipx", "venvs", binaryName)
	info, err := os.Stat(venvPath)
	return err == nil && info.IsDir()
}
