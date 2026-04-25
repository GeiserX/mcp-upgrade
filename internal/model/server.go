package model

type ServerType string

const (
	TypeNPX           ServerType = "npx"
	TypeLocalNode     ServerType = "local-node"
	TypePipx          ServerType = "pipx"
	TypeUVX           ServerType = "uvx"
	TypeDocker        ServerType = "docker"
	TypeGoBinary      ServerType = "go-binary"
	TypeGitHubRelease ServerType = "github-release"
	TypeCargo         ServerType = "cargo"
	TypeLocal         ServerType = "local"
	TypeUnknown       ServerType = "unknown"
)

type UpgradeStatus string

const (
	StatusUpToDate   UpgradeStatus = "UP-TO-DATE"
	StatusUpgradable UpgradeStatus = "UPDATE"
	StatusAutoLatest UpgradeStatus = "AUTO"
	StatusSkipped    UpgradeStatus = "SKIP"
	StatusUnknown    UpgradeStatus = "UNKNOWN"
	StatusError      UpgradeStatus = "ERROR"
)

type Client struct {
	Name       string
	ConfigPath string
}

type Server struct {
	Name           string     `json:"name"`
	Command        string     `json:"command"`
	Args           []string   `json:"args"`
	Env            map[string]string `json:"env,omitempty"`
	Type           ServerType `json:"type"`
	Package        string     `json:"package"`
	CurrentVersion string     `json:"current_version"`
	LatestVersion  string     `json:"latest_version"`
	Status         UpgradeStatus `json:"status"`
	Client         string     `json:"client"`
	ConfigPath     string     `json:"config_path"`
	GitHubRepo     string     `json:"github_repo,omitempty"`
	BinaryPath     string     `json:"binary_path,omitempty"`
	DockerImage    string     `json:"docker_image,omitempty"`
}

type ScanResult struct {
	Servers []Server `json:"servers"`
	Clients []Client `json:"clients"`
}
