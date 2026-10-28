package automation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	unityPackageManifestKey = "Packages/src"
	unityPackageName        = "io.github.tongguang.tloop"
)

// unityPackageReleaseTagPattern accepts only the Unity package component's tags. The CLI
// components share the repository and its releases, so their tags must never select the package
// source for packing.
var unityPackageReleaseTagPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$`)

// releaseCommitPattern matches a full commit SHA. The package release is packed from, and tagged
// at, its release commit, so anything that could resolve to another commit later (a branch, an
// abbreviation) is refused.
var releaseCommitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// unityPackageAssetName is the release asset OpenUPM picks up; its fixed prefix is what the
// OpenUPM package metadata matches with githubReleaseAssetName.
func unityPackageAssetName(version string) string {
	return unityPackageName + "-" + version + ".tgz"
}

func readUnityPackageManifestVersion(repoRoot string) (string, error) {
	content, err := os.ReadFile(filepath.Join(repoRoot, releasePleaseManifestRelativePath))
	if err != nil {
		return "", fmt.Errorf("read release-please manifest: %w", err)
	}
	var manifest map[string]string
	if err := json.Unmarshal(content, &manifest); err != nil {
		return "", fmt.Errorf("parse release-please manifest: %w", err)
	}
	version := manifest[unityPackageManifestKey]
	if version == "" {
		return "", fmt.Errorf("release-please manifest has no version for %s", unityPackageManifestKey)
	}
	return version, nil
}
