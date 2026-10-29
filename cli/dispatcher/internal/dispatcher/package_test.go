package dispatcher

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Verifies package help lists install, status, and --version.
func TestPackageHelpListsSubcommands(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunDispatcher(context.Background(), []string{"package", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("package help failed: code=%d stderr=%s", code, stderr.String())
	}
	output := stdout.String()
	for _, expected := range []string{"install", "status", "--version", "--project-path"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("package help missing %q:\n%s", expected, output)
		}
	}
}

// Verifies the default install writes the fork Git URL without requiring a registry.
func TestPackageInstallWritesGitDependency(t *testing.T) {
	projectRoot := createPackageTestProject(t, barePackageManifest())
	code, stdout, stderr := runPackageForTest(t, projectRoot, "install")
	if code != 0 {
		t.Fatalf("install failed: code=%d stderr=%s", code, stderr)
	}
	content := readPackageManifest(t, projectRoot)
	assertManifestHasDependency(t, []byte(content), dispatcherUnityPackageName, "https://github.com/tongguang/unity-cli-loop.git?path=/Packages/src")
	if strings.Contains(content, "scopedRegistries") || strings.Contains(stdout, "OpenUPM") {
		t.Fatalf("install must use only the Git dependency: manifest=%s stdout=%s", content, stdout)
	}
}

// Verifies release versions, with or without a v prefix, become pinned Git dependencies.
func TestPackageInstallPinsGitReleaseTag(t *testing.T) {
	cases := []struct{ version, tag string }{
		{"1.2.3", "v1.2.3"},
		{"v1.2.3", "v1.2.3"},
		{"V1.2.3", "v1.2.3"},
		{"1.2.3-beta.1", "v1.2.3-beta.1"},
	}
	for _, tc := range cases {
		t.Run(tc.version, func(t *testing.T) {
			projectRoot := createPackageTestProject(t, barePackageManifest())
			code, _, stderr := runPackageForTest(t, projectRoot, "install", "--version", tc.version)
			if code != 0 {
				t.Fatalf("install failed: code=%d stderr=%s", code, stderr)
			}
			assertManifestHasDependency(t, []byte(readPackageManifest(t, projectRoot)), dispatcherUnityPackageName, packageGitURL+"#"+tc.tag)
		})
	}
}

// Verifies Git installation preserves other package dependencies and scoped registries.
func TestPackageInstallPreservesOtherPackagesAndRegistries(t *testing.T) {
	manifest := `{
  "dependencies": {"io.github.hatayama.uloopmcp": "3.21.1"},
  "scopedRegistries": [{"name":"OpenUPM","url":"https://package.openupm.com","scopes":["io.github.hatayama.uloopmcp"]}]
}
`
	projectRoot := createPackageTestProject(t, manifest)
	code, _, stderr := runPackageForTest(t, projectRoot, "install", "--version", "3.22.2")
	if code != 0 {
		t.Fatalf("install failed: code=%d stderr=%s", code, stderr)
	}
	content := []byte(readPackageManifest(t, projectRoot))
	assertManifestHasDependency(t, content, "io.github.hatayama.uloopmcp", "3.21.1")
	assertManifestHasDependency(t, content, dispatcherUnityPackageName, packageGitURL+"#v3.22.2")
	assertManifestRegistriesUnchanged(t, []byte(manifest), content)
}

// Verifies a second install reports already installed and leaves the manifest unchanged.
func TestPackageInstallIsIdempotent(t *testing.T) {
	projectRoot := createPackageTestProject(t, barePackageManifest())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	firstCode := RunDispatcher(
		context.Background(),
		[]string{"package", "install", "--project-path", projectRoot},
		&stdout,
		&stderr,
	)
	if firstCode != 0 {
		t.Fatalf("first install failed: code=%d stderr=%s", firstCode, stderr.String())
	}
	before := readPackageManifest(t, projectRoot)

	stdout.Reset()
	stderr.Reset()
	secondCode := RunDispatcher(
		context.Background(),
		[]string{"package", "install", "--project-path", projectRoot},
		&stdout,
		&stderr,
	)
	if secondCode != 0 {
		t.Fatalf("second install failed: code=%d stderr=%s", secondCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "already installed") {
		t.Fatalf("expected already installed message:\n%s", stdout.String())
	}
	after := readPackageManifest(t, projectRoot)
	if before != after {
		t.Fatalf("manifest changed on idempotent install:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// Verifies a project without Packages/manifest.json returns PACKAGE_MANIFEST_INVALID.
func TestPackageInstallFailsWithoutManifest(t *testing.T) {
	projectRoot := t.TempDir()
	for _, directory := range []string{"Assets", "ProjectSettings", "Packages"} {
		if err := os.MkdirAll(filepath.Join(projectRoot, directory), 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunDispatcher(
		context.Background(),
		[]string{"package", "install", "--version", "1.2.3", "--project-path", projectRoot},
		&stdout,
		&stderr,
	)
	if code == 0 {
		t.Fatal("expected failure without manifest")
	}
	if !strings.Contains(stderr.String(), "PACKAGE_MANIFEST_INVALID") {
		t.Fatalf("expected PACKAGE_MANIFEST_INVALID:\n%s", stderr.String())
	}
}

// Verifies status reports not installed when the package is absent.
func TestPackageStatusReportsNotInstalled(t *testing.T) {
	projectRoot := createPackageTestProject(t, barePackageManifest())
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunDispatcher(
		context.Background(),
		[]string{"package", "status", "--project-path", projectRoot},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("status failed: code=%d stderr=%s", code, stderr.String())
	}
	output := stdout.String()
	for _, expected := range []string{
		"Package: " + dispatcherUnityPackageName,
		"Dependency: not installed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("status output missing %q:\n%s", expected, output)
		}
	}
}

// Verifies status reports a Git dependency without requiring a scoped registry.
func TestPackageStatusReportsInstalledVersion(t *testing.T) {
	dependency := packageGitURL + "#v1.2.3"
	projectRoot := createPackageTestProject(t, installedPackageManifest(dependency))
	code, stdout, stderr := runPackageForTest(t, projectRoot, "status")
	if code != 0 || !strings.Contains(stdout, "Dependency: installed ("+dependency+")") {
		t.Fatalf("unexpected status: code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	if strings.Contains(stdout, "Scoped registry") {
		t.Fatalf("unexpected registry requirement: %s", stdout)
	}
}

// Verifies status reads package dependencies independently of unrelated registry settings.
func TestInspectPackageManifestStatusCharacterizesInstalledFlags(t *testing.T) {
	cases := []struct {
		name, content string
		want          packageManifestStatus
	}{
		{"bare manifest", barePackageManifest(), packageManifestStatus{}},
		{"Git dependency", installedPackageManifest(packageGitURL), packageManifestStatus{dependencyInstalled: true, dependencyVersion: packageGitURL}},
		{"pinned Git dependency", installedPackageManifest(packageGitURL + "#v1.2.3"), packageManifestStatus{dependencyInstalled: true, dependencyVersion: packageGitURL + "#v1.2.3"}},
		{"registry without dependency", `{"scopedRegistries":[{"url":"https://package.openupm.com"}]}`, packageManifestStatus{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, err := inspectPackageManifestStatus([]byte(tc.content))
			if err != nil || status != tc.want {
				t.Fatalf("status=%+v want=%+v err=%v", status, tc.want, err)
			}
		})
	}
}

// Verifies an unknown package subcommand fails with guidance for valid subcommands.
func TestPackageUnknownSubcommandFails(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunDispatcher(context.Background(), []string{"package", "bogus"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "tloop package install") || !strings.Contains(stderr.String(), "tloop package status") {
		t.Fatalf("expected valid subcommand guidance:\n%s", stderr.String())
	}
}

func createPackageTestProject(t *testing.T, manifest string) string {
	t.Helper()
	projectRoot := t.TempDir()
	createPackageProjectAt(t, projectRoot, manifest)
	return projectRoot
}

func barePackageManifest() string {
	return `{
  "dependencies": {
    "com.unity.modules.ai": "1.0.0"
  }
}
`
}

func installedPackageManifest(dependency string) string {
	return `{
  "dependencies": {
    "com.unity.modules.ai": "1.0.0",
    "io.github.tongguang.tloop": "` + dependency + `"
  }
}
`
}

func readPackageManifest(t *testing.T, projectRoot string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(projectRoot, "Packages", "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}
	return string(content)
}

// Verifies install resolves a Unity project in a child directory when run from a repository root.
func TestPackageInstallFindsProjectInChildDirectory(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	projectRoot := filepath.Join(repoRoot, "client", "UnityApp")
	createPackageProjectAt(t, projectRoot, barePackageManifest())
	t.Chdir(repoRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunDispatcher(
		context.Background(),
		[]string{"package", "install", "--version", "9.9.9"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("package install failed: code=%d stderr=%s", code, stderr.String())
	}

	content := readPackageManifest(t, projectRoot)
	if !strings.Contains(content, `"`+dispatcherUnityPackageName+`": "`+packageGitURL+`#v9.9.9"`) {
		t.Fatalf("dependency missing from manifest:\n%s", content)
	}
}

// Verifies install fails with guidance when multiple Unity projects exist under the start directory.
func TestPackageInstallFailsWithMultipleChildProjects(t *testing.T) {
	repoRoot := t.TempDir()
	createPackageProjectAt(t, filepath.Join(repoRoot, "AppA"), barePackageManifest())
	createPackageProjectAt(t, filepath.Join(repoRoot, "AppB"), barePackageManifest())
	t.Chdir(repoRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunDispatcher(
		context.Background(),
		[]string{"package", "install", "--version", "9.9.9"},
		&stdout,
		&stderr,
	)
	if code == 0 {
		t.Fatalf("package install must fail with multiple candidate projects: stdout=%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "--project-path") {
		t.Fatalf("error must guide toward --project-path:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), `"ErrorCode": "PROJECT_NOT_FOUND"`) {
		t.Fatalf("ambiguous resolution must classify as PROJECT_NOT_FOUND, not an internal error:\n%s", stderr.String())
	}
}

// Verifies install prefers the enclosing Unity project over a nested Unity-shaped child folder
// when run from a subdirectory of a Unity project.
func TestPackageInstallPrefersEnclosingProjectOverNestedChild(t *testing.T) {
	projectRoot := createPackageTestProject(t, barePackageManifest())
	nestedFixture := filepath.Join(projectRoot, "ci", "fixtures", "NestedApp")
	createPackageProjectAt(t, nestedFixture, barePackageManifest())
	t.Chdir(filepath.Join(projectRoot, "ci"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunDispatcher(
		context.Background(),
		[]string{"package", "install", "--version", "9.9.9"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("package install failed: code=%d stderr=%s", code, stderr.String())
	}

	if !strings.Contains(readPackageManifest(t, projectRoot), `"`+dispatcherUnityPackageName+`": "`+packageGitURL+`#v9.9.9"`) {
		t.Fatalf("enclosing project manifest must be updated")
	}
	if strings.Contains(readPackageManifest(t, nestedFixture), dispatcherUnityPackageName) {
		t.Fatalf("nested fixture manifest must stay untouched")
	}
}

// Verifies status resolves a Unity project in a child directory when run from a repository root.
func TestPackageStatusFindsProjectInChildDirectory(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	createPackageProjectAt(t, filepath.Join(repoRoot, "UnityApp"), barePackageManifest())
	t.Chdir(repoRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunDispatcher(context.Background(), []string{"package", "status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("package status failed: code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Dependency: not installed") {
		t.Fatalf("status output unexpected:\n%s", stdout.String())
	}
}

func createPackageProjectAt(t *testing.T, projectRoot string, manifest string) {
	t.Helper()
	for _, directory := range []string{"Assets", "ProjectSettings", "Packages"} {
		if err := os.MkdirAll(filepath.Join(projectRoot, directory), 0o755); err != nil {
			t.Fatalf("failed to create %s: %v", directory, err)
		}
	}
	manifestPath := filepath.Join(projectRoot, "Packages", "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
}

// Verifies skills install resolves a Unity project in a child directory the same way package install does.
func TestResolveSkillsProjectRootFindsProjectInChildDirectory(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	projectRoot := filepath.Join(repoRoot, "UnityApp")
	createPackageProjectAt(t, projectRoot, barePackageManifest())

	resolved, err := resolveSkillsProjectRoot(repoRoot, "", false)
	if err != nil {
		t.Fatalf("resolveSkillsProjectRoot failed: %v", err)
	}
	if resolved != projectRoot {
		t.Fatalf("project root mismatch: %s", resolved)
	}
}

func TestParsePackageOptionsRejectInvalidArguments(t *testing.T) {
	// Verifies each invalid install or status argument is rejected with the message for that mistake.
	installCases := []struct {
		args        []string
		wantMessage string
	}{
		{args: []string{"--bogus", "1"}, wantMessage: "Unknown package install option: --bogus"},
		{args: []string{"--version", ""}, wantMessage: "Empty value for --version"},
		{args: []string{"--version", "1.0.0", "--version", "2.0.0"}, wantMessage: "Duplicate package install option: --version"},
		{args: []string{"--version"}, wantMessage: "--version requires a value"},
	}
	for _, testCase := range installCases {
		if _, err := parsePackageInstallOptions(testCase.args); err == nil || !strings.Contains(err.Error(), testCase.wantMessage) {
			t.Fatalf("%v: expected %q, got %v", testCase.args, testCase.wantMessage, err)
		}
	}
	if err := parsePackageStatusOptions([]string{"--verbose"}); err == nil || err.Error() != "Unknown package status option: --verbose" {
		t.Fatalf("expected the status option error, got %v", err)
	}
}

func TestTryHandlePackageRequestPrintsSubcommandHelp(t *testing.T) {
	// Verifies --help after a subcommand prints package help without touching any project.
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	handled, code := tryHandlePackageRequest(context.Background(), []string{"package", "install", "--help"}, t.TempDir(), "", &stdout, &stderr)

	if !handled || code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "tloop package install [--version <x.y.z>]") {
		t.Fatalf("unexpected help: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

// runPackageForTest runs a package subcommand against an explicit project and returns its exit code and output.
func runPackageForTest(t *testing.T, projectRoot string, args ...string) (int, string, string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	_, code := tryHandlePackageRequest(context.Background(), append([]string{"package"}, args...), t.TempDir(), projectRoot, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestPackageCommandsReportFailures(t *testing.T) {
	// Verifies install and status stop with code 1 and the error for each failure, leaving the manifest untouched.
	cases := []struct {
		name        string
		manifest    string
		args        []string
		wantMessage string
	}{
		{name: "install invalid option", manifest: barePackageManifest(), args: []string{"install", "--bogus=1"}, wantMessage: "Unknown package install option: --bogus"},
		{name: "install invalid manifest", manifest: "{", args: []string{"install", "--version", "1.0.0"}, wantMessage: "PACKAGE_MANIFEST_INVALID"},
		{name: "status invalid option", manifest: barePackageManifest(), args: []string{"status", "--verbose"}, wantMessage: "Unknown package status option: --verbose"},
		{name: "status invalid manifest", manifest: "{", args: []string{"status"}, wantMessage: "PACKAGE_MANIFEST_INVALID"},
		{name: "status dependency not a string", manifest: `{"dependencies":{"` + dispatcherUnityPackageName + `":1}}`, args: []string{"status"}, wantMessage: "PACKAGE_MANIFEST_INVALID"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			projectRoot := createPackageTestProject(t, testCase.manifest)

			code, stdout, stderr := runPackageForTest(t, projectRoot, testCase.args...)

			if code != 1 || stdout != "" || !strings.Contains(stderr, testCase.wantMessage) {
				t.Fatalf("unexpected result: code=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
			if manifest := readPackageManifest(t, projectRoot); manifest != testCase.manifest {
				t.Fatalf("manifest must stay untouched: %s", manifest)
			}
		})
	}
}

func TestPackageStatusReportsMissingManifestAndProject(t *testing.T) {
	// Verifies status distinguishes a project without Packages/manifest.json from a path that is no Unity project.
	projectRoot := createDispatcherUnityProject(t)
	code, _, stderr := runPackageForTest(t, projectRoot, "status")
	if code != 1 || !strings.Contains(stderr, "PACKAGE_MANIFEST_INVALID") {
		t.Fatalf("expected the manifest error: code=%d stderr=%s", code, stderr)
	}

	code, _, stderr = runPackageForTest(t, t.TempDir(), "status")
	if code != 1 || strings.Contains(stderr, "PACKAGE_MANIFEST_INVALID") || !strings.Contains(stderr, "Unity project") {
		t.Fatalf("expected the project resolution error: code=%d stderr=%s", code, stderr)
	}
}

// Verifies invalid Git version inputs fail without replacing an existing working dependency.
func TestPackageInstallRejectsInvalidVersionsWithoutChangingManifest(t *testing.T) {
	for _, value := range []string{"main", "1.2", "1.2.3#main", "1.2.3/../main", " 1.2.3 "} {
		t.Run(value, func(t *testing.T) {
			manifest := installedPackageManifest(packageGitURL + "#v3.22.2")
			projectRoot := createPackageTestProject(t, manifest)
			code, _, stderr := runPackageForTest(t, projectRoot, "install", "--version", value)
			if code != 1 || !strings.Contains(stderr, "Invalid package version") {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			if after := readPackageManifest(t, projectRoot); after != manifest {
				t.Fatalf("manifest changed: %s", after)
			}
		})
	}
}

func TestPackageInstallReportsManifestWriteFailure(t *testing.T) {
	// Verifies an unwritable Packages directory fails the install and keeps the original manifest.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory permissions are required for this failure.")
	}
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions.")
	}
	projectRoot := createPackageTestProject(t, barePackageManifest())
	packagesDir := filepath.Join(projectRoot, "Packages")
	if err := os.Chmod(packagesDir, 0o555); err != nil {
		t.Fatalf("failed to lock %s: %v", packagesDir, err)
	}
	t.Cleanup(func() { _ = os.Chmod(packagesDir, 0o755) })

	code, _, stderr := runPackageForTest(t, projectRoot, "install", "--version", "1.0.0")

	if code != 1 || !strings.Contains(stderr, "manifest.json.tmp: permission denied") {
		t.Fatalf("expected the write error: code=%d stderr=%s", code, stderr)
	}
	if manifest := readPackageManifest(t, projectRoot); manifest != barePackageManifest() {
		t.Fatalf("manifest must stay untouched: %s", manifest)
	}
}

// Verifies upgrading changes a Git tag and reports both dependency URLs.
func TestPackageInstallReportsVersionUpgrade(t *testing.T) {
	previous := packageGitURL + "#v1.0.0"
	next := packageGitURL + "#v2.0.0"
	projectRoot := createPackageTestProject(t, installedPackageManifest(previous))
	code, stdout, stderr := runPackageForTest(t, projectRoot, "install", "--version", "2.0.0")
	if code != 0 || !strings.Contains(stdout, "Updated "+dispatcherUnityPackageName+" "+previous+" -> "+next+" in Packages/manifest.json") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	assertManifestHasDependency(t, []byte(readPackageManifest(t, projectRoot)), dispatcherUnityPackageName, next)
}

// Verifies malformed package dependencies fail inspection instead of appearing absent.
func TestInspectPackageManifestStatusRejectsMalformedEntries(t *testing.T) {
	for _, manifest := range []string{`{not json`, `{"dependencies":[]}`, `{"dependencies":{"io.github.tongguang.tloop":1}}`} {
		if _, err := inspectPackageManifestStatus([]byte(manifest)); err == nil {
			t.Fatalf("expected an error for %s", manifest)
		}
	}
}

// Verifies reinstalling a pinned Git dependency never replaces it with a registry version.
func TestPackageInstallPreservesMatchingGitDependency(t *testing.T) {
	manifest := installedPackageManifest(packageGitURL + "#v3.22.2")
	projectRoot := createPackageTestProject(t, manifest)
	code, stdout, stderr := runPackageForTest(t, projectRoot, "install", "--version", "3.22.2")
	if code != 0 || !strings.Contains(stdout, "already installed") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	if after := readPackageManifest(t, projectRoot); after != manifest {
		t.Fatalf("manifest changed: %s", after)
	}
}
