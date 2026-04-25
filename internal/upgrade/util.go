package upgrade

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// execCommand runs a command and returns its combined stdout+stderr output.
func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// findAsset searches a list of release asset names for one matching the given
// OS and architecture. It handles common naming conventions like Darwin_arm64,
// darwin-arm64, linux_amd64, Linux-x86_64, etc.
func findAsset(assets []string, goos, goarch string) string {
	osAliases := osPatterns(goos)
	archAliases := archPatterns(goarch)

	for _, asset := range assets {
		lower := strings.ToLower(asset)
		if !hasAny(lower, osAliases) {
			continue
		}
		if !hasAny(lower, archAliases) {
			continue
		}
		// Skip checksum and signature files.
		if strings.HasSuffix(lower, ".sha256") ||
			strings.HasSuffix(lower, ".sha512") ||
			strings.HasSuffix(lower, ".sig") ||
			strings.HasSuffix(lower, ".asc") ||
			strings.HasSuffix(lower, ".sbom") {
			continue
		}
		return asset
	}
	return ""
}

func osPatterns(goos string) []string {
	switch goos {
	case "darwin":
		return []string{"darwin", "macos", "apple"}
	case "linux":
		return []string{"linux"}
	case "windows":
		return []string{"windows", "win64", "win32"}
	default:
		return []string{goos}
	}
}

func archPatterns(goarch string) []string {
	switch goarch {
	case "amd64":
		return []string{"amd64", "x86_64", "x64"}
	case "arm64":
		return []string{"arm64", "aarch64"}
	case "386":
		return []string{"386", "i386", "i686", "x86"}
	default:
		return []string{goarch}
	}
}

func hasAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// extractArchive extracts .tar.gz or .zip archives into destDir.
func extractArchive(path, destDir string) error {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(path, destDir)
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(path, destDir)
	default:
		return fmt.Errorf("unsupported archive format: %s", filepath.Base(path))
	}
}

func extractTarGz(path, destDir string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, filepath.Clean(hdr.Name))
		// Prevent directory traversal.
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != filepath.Clean(destDir) {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
	return nil
}

func extractZip(path, destDir string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, filepath.Clean(f.Name))
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != filepath.Clean(destDir) {
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, copyErr := io.Copy(out, rc)
		out.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}
