package dispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Verifies Tuanjie versions use Tuanjie Hub paths on both Windows and Linux.
func TestTuanjieExecutableCandidates(t *testing.T) {
	t.Setenv("ProgramFiles", "<PROGRAM_FILES>")
	t.Setenv("ProgramFiles(x86)", "")
	t.Setenv("LOCALAPPDATA", "")
	want := []string{
		filepath.Join("<PROGRAM_FILES>", "Tuanjie", "Hub", "Editor", "2022.3.62t15", "Editor", "Tuanjie.exe"),
		filepath.Join(`C:\Program Files`, "Tuanjie", "Hub", "Editor", "2022.3.62t15", "Editor", "Tuanjie.exe"),
	}
	if got := windowsUnityExecutableCandidates("2022.3.62t15"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Windows candidates = %v, want %v", got, want)
	}
	if got := linuxUnityExecutableCandidates("2022.3.62t15", "/home/user"); !reflect.DeepEqual(got, []string{
		"/home/user/Tuanjie/Hub/Editor/2022.3.62t15/Editor/Tuanjie",
	}) {
		t.Fatalf("Linux candidates = %v", got)
	}
}

// Verifies BOM, CRLF and a Tuanjie release suffix survive project version parsing.
func TestReadTuanjieEditorVersionOnWindows(t *testing.T) {
	projectRoot := t.TempDir()
	writeDispatcherTestFile(t, filepath.Join(projectRoot, projectVersionFilePath),
		"\ufeffm_EditorVersion: 2022.3.62t15\r\nm_EditorVersionWithRevision: 2022.3.62t15 (abc)\r\n")
	got, err := readUnityEditorVersion(projectRoot)
	if err != nil || got != "2022.3.62t15" {
		t.Fatalf("version = %q, err = %v", got, err)
	}
}

// Verifies an empty version cannot consume the following revision line.
func TestReadEditorVersionDoesNotReadNextLine(t *testing.T) {
	projectRoot := t.TempDir()
	writeDispatcherTestFile(t, filepath.Join(projectRoot, projectVersionFilePath),
		"m_EditorVersion: \r\nm_EditorVersionWithRevision: 2022.3.62t15 (abc)\r\n")
	if _, err := readUnityEditorVersion(projectRoot); err == nil || !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("expected an empty version error, got %v", err)
	}
}

// Verifies Hub metadata finds custom installs, preserves Unicode paths and expands macOS bundles.
func TestHubEditorExecutableCandidates(t *testing.T) {
	version := "2022.3.62t15"
	configDir := t.TempDir()
	customPath := filepath.Join(t.TempDir(), "团结 引擎", "Tuanjie.exe")
	bundlePath := filepath.Join(t.TempDir(), "Tuanjie.app")
	content, err := json.Marshal(map[string]any{"schema_version": "2", "data": []any{
		map[string]any{"version": "other", "location": []string{"wrong"}},
		map[string]any{"version": version, "location": []string{customPath, bundlePath}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	writeDispatcherTestFile(t, filepath.Join(configDir, "TuanjieHub", "editors-v2.json"), "\ufeff"+string(content)+"\r\n")
	want := []string{customPath, filepath.Join(bundlePath, "Contents", "MacOS", "Tuanjie")}
	if got := hubEditorExecutableCandidates(version, configDir); !reflect.DeepEqual(got, want) {
		t.Fatalf("candidates = %v, want %v", got, want)
	}
}

// Verifies optional Hub metadata never prevents the default install search.
func TestHubEditorExecutableCandidatesIgnoreMissingOrInvalidMetadata(t *testing.T) {
	configDir := t.TempDir()
	for _, directory := range []string{"", configDir} {
		if got := hubEditorExecutableCandidates("2022.3.62t15", directory); len(got) != 0 {
			t.Fatalf("unexpected candidates: %v", got)
		}
	}
	for _, content := range []string{"not JSON", `{"data":[]}`, `{"data":[{"version":"other","location":["wrong"]}]}`} {
		writeDispatcherTestFile(t, filepath.Join(configDir, "TuanjieHub", "editors-v2.json"), content)
		if got := hubEditorExecutableCandidates("2022.3.62t15", configDir); len(got) != 0 {
			t.Fatalf("unexpected candidates for %q: %v", content, got)
		}
	}
}

// Verifies both native commands accept a custom Editor and reject ambiguous or incomplete choices.
func TestParseEditorPathOptions(t *testing.T) {
	for _, args := range [][]string{{"--editor-path", "custom editor"}, {"--editor-path=custom editor"}} {
		launch, err := parseLaunchOptions(args, "")
		if err != nil || launch.editorPath != "custom editor" {
			t.Fatalf("launch options = %+v, err = %v", launch, err)
		}
		check, err := parseCompileCheckOptions(args, "")
		if err != nil || check.editorPath != "custom editor" {
			t.Fatalf("compile-check options = %+v, err = %v", check, err)
		}
	}
	for _, args := range [][]string{
		{"--editor-path"},
		{"--editor-path="},
		{"--editor-path", "--all"},
		{"--editor-path=custom", "--editor-version=2022.3.62t15"},
	} {
		if _, err := parseLaunchOptions(args, ""); err == nil {
			t.Fatalf("launch accepted %v", args)
		}
		if _, err := parseCompileCheckOptions(args, ""); err == nil {
			t.Fatalf("compile-check accepted %v", args)
		}
	}
}

// Verifies an explicit executable bypasses version discovery for launch and offline compilation.
func TestExplicitEditorPathBypassesDiscovery(t *testing.T) {
	editorPath := filepath.Join(t.TempDir(), "团结 引擎", "Tuanjie.exe")
	writeDispatcherTestFile(t, editorPath, "")
	deps := isolatedLaunchTestDeps(t)
	deps.resolveUnityExecutablePath = func(string) (string, error) {
		t.Fatal("an explicit Editor path must bypass version discovery")
		return "", nil
	}
	options := launchOptions{editorPath: editorPath}
	version, err := resolveLaunchEditorVersion(t.TempDir(), options)
	if err != nil || version != "" {
		t.Fatalf("version = %q, err = %v", version, err)
	}
	got, err := resolveLaunchExecutablePath(options, version, deps)
	if err != nil || got != editorPath {
		t.Fatalf("launch path = %q, err = %v", got, err)
	}
	got, err = resolveCompileCheckEditorPath(t.TempDir(), compileCheckOptions{editorPath: editorPath})
	if err != nil || got != editorPath {
		t.Fatalf("compile-check path = %q, err = %v", got, err)
	}
}

// Verifies an Editor override cannot silently reuse an already running Editor.
func TestExistingEditorPathRequiresRestart(t *testing.T) {
	deps := isolatedLaunchTestDeps(t)
	var stdout, stderr bytes.Buffer
	_, handled, code := handleExistingLaunchProcess(context.Background(), launchOptions{editorPath: "custom"},
		t.TempDir(), nil, false, &stdout, &stderr, deps)
	if !handled || code != 1 || !strings.Contains(stderr.String(), "--editor-path requires --restart") {
		t.Fatalf("handled = %v, code = %d, stderr = %s", handled, code, &stderr)
	}
}

// Verifies discovery skips directories and explicit selection rejects missing files and directories.
func TestEditorExecutableMustBeAFile(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "Tuanjie.exe")
	writeDispatcherTestFile(t, executable, "")
	got, err := resolveExistingUnityExecutablePath("2022.3.62t15", []string{directory, executable})
	if err != nil || got != executable {
		t.Fatalf("path = %q, err = %v", got, err)
	}
	for _, invalid := range []string{directory, filepath.Join(directory, "missing.exe")} {
		if _, err := resolveExplicitEditorPath(invalid); err == nil {
			t.Fatalf("accepted %q", invalid)
		}
	}
}
