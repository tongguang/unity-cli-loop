package dispatcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var editorVersionPattern = regexp.MustCompile(`(?m)^m_EditorVersion:[\t ]*([^\r\n]*)`)

func resolveUnityExecutablePath(version string) (string, error) {
	configDir, _ := os.UserConfigDir()
	candidates := append(hubEditorExecutableCandidates(version, configDir), unityExecutableCandidates(version)...)
	return resolveExistingUnityExecutablePath(version, candidates)
}

func resolveLaunchExecutablePath(options launchOptions, version string, deps launchDeps) (string, error) {
	if options.editorPath != "" {
		return resolveExplicitEditorPath(options.editorPath)
	}
	return deps.resolveUnityExecutablePath(version)
}

func resolveLaunchEditorVersion(projectRoot string, options launchOptions) (string, error) {
	if options.editorPath != "" {
		return "", nil
	}
	if options.editorVersion != "" {
		return options.editorVersion, nil
	}
	return readUnityEditorVersion(projectRoot)
}

func resolveExistingUnityExecutablePath(version string, candidates []string) (string, error) {
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("tuanjie launch is not supported on %s", runtime.GOOS)
	}
	return "", fmt.Errorf("tuanjie %s executable not found; searched: %s; use --editor-path for a custom installation", version, strings.Join(candidates, ", "))
}

func unityExecutableCandidates(version string) []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{fmt.Sprintf("/Applications/Tuanjie/Hub/Editor/%s/Tuanjie.app/Contents/MacOS/Tuanjie", version)}
	case "windows":
		return windowsUnityExecutableCandidates(version)
	case "linux":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return []string{}
		}
		return linuxUnityExecutableCandidates(version, homeDir)
	default:
		return []string{}
	}
}

func windowsUnityExecutableCandidates(version string) []string {
	candidates := []string{}
	for _, base := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("LOCALAPPDATA"), `C:\Program Files`} {
		if base != "" {
			candidates = append(candidates, filepath.Join(base, "Tuanjie", "Hub", "Editor", version, "Editor", "Tuanjie.exe"))
		}
	}
	return candidates
}

// linuxUnityExecutableCandidates lists where Tuanjie Hub installs an Editor on Linux.
// An unknown home directory yields no candidate rather than a relative path that
// would be resolved against the caller's working directory. path.Join keeps the
// '/'-separated result identical when the tests run on Windows.
func linuxUnityExecutableCandidates(version string, homeDir string) []string {
	if homeDir == "" {
		return []string{}
	}
	return []string{path.Join(homeDir, "Tuanjie", "Hub", "Editor", version, "Editor", "Tuanjie")}
}

// Hub records manually located Editors here, including non-default install directories.
func hubEditorExecutableCandidates(version string, configDir string) []string {
	if configDir == "" {
		return nil
	}
	content, err := os.ReadFile(filepath.Join(configDir, "TuanjieHub", "editors-v2.json"))
	if err != nil {
		return nil
	}
	var installed struct {
		Data []struct {
			Version  string   `json:"version"`
			Location []string `json:"location"`
		} `json:"data"`
	}
	if json.Unmarshal(bytes.TrimPrefix(content, []byte{0xef, 0xbb, 0xbf}), &installed) != nil {
		return nil
	}
	var candidates []string
	for _, editor := range installed.Data {
		if editor.Version == version {
			for _, location := range editor.Location {
				if strings.HasSuffix(location, ".app") {
					location = filepath.Join(location, "Contents", "MacOS", "Tuanjie")
				}
				candidates = append(candidates, location)
			}
		}
	}
	return candidates
}

func resolveExplicitEditorPath(editorPath string) (string, error) {
	absolutePath, err := filepath.Abs(editorPath)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("editor executable not found: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("editor executable path is a directory: %s", absolutePath)
	}
	return absolutePath, nil
}

func readUnityEditorVersion(projectRoot string) (string, error) {
	content, err := os.ReadFile(filepath.Join(projectRoot, projectVersionFilePath))
	if err != nil {
		return "", err
	}
	matches := editorVersionPattern.FindStringSubmatch(string(bytes.TrimPrefix(content, []byte{0xef, 0xbb, 0xbf})))
	if len(matches) != 2 {
		return "", fmt.Errorf("tuanjie editor version not found in %s", projectVersionFilePath)
	}
	version := strings.TrimSpace(matches[1])
	if version == "" {
		return "", fmt.Errorf("tuanjie editor version is empty in %s", projectVersionFilePath)
	}
	return version, nil
}
