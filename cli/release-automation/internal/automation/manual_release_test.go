package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

type manualReleaseFixture struct {
	root    string
	head    string
	remote  string
	calls   []string
	changed map[string]bool
	tags    map[string]string
	current manualReleaseWorkflowRun
	fail    string
	count   int
}

func newManualReleaseFixture(t *testing.T) *manualReleaseFixture {
	t.Helper()
	fixture := &manualReleaseFixture{root: t.TempDir(), head: "source", remote: "source", changed: map[string]bool{}, tags: map[string]string{}}
	files := map[string]string{
		releasePleaseManifestRelativePath:                            `{"cli/dispatcher":"3.14.2","cli/project-runner":"3.13.1","Packages/src":"3.22.3"}`,
		"cli/dispatcher/dispatchercontract/dispatcher-contract.json": `{"dispatcherVersion":"3.14.2","protocolVersion":5}`,
		"cli/common/clicontract/contract.json":                       `{"projectRunnerVersion":"3.13.1","protocolVersion":5}`,
		"Packages/src/package.json":                                  `{"version":"3.22.3","name":"io.github.tongguang.tloop"}`,
		unityPackageCliPinFile:                                       `{"projectRunnerVersion":"3.13.1","minimumDispatcherVersion":"3.14.0","dispatcherReleaseTag":"dispatcher-v3.14.2","future":{"保留":true}}`,
		".tloop/project-runner-pin.json":                             `{}`,
	}
	for name, content := range files {
		writeManualFixtureFile(t, fixture.root, name, content)
	}
	for _, component := range manualReleaseComponentDefinitions() {
		writeManualFixtureFile(t, fixture.root, component.Root+"/CHANGELOG.md", "# Changelog\r\n\r\n## [3.0.0]\r\n\r\n- Existing history.\r\n")
	}
	return fixture
}

func writeManualFixtureFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (fixture *manualReleaseFixture) options() manualReleaseOptions {
	return manualReleaseOptions{Root: fixture.root, Repository: "tongguang/unity-cli-loop", Bump: "patch", Components: "auto", RunID: "123-1"}
}

func (fixture *manualReleaseFixture) deps() manualReleaseDeps {
	return manualReleaseDeps{
		run: fixture.run,
		wait: func(context.Context, time.Duration) error {
			return nil
		},
		stamp: func(_ context.Context, pin, tag string) error {
			fixture.calls = append(fixture.calls, "stamp "+tag)
			if fixture.fail == "stamp" {
				return errors.New("invalid attestation")
			}
			return updateManualReleaseJSON(fixture.root, unityPackageCliPinFile, "dispatcherReleaseTag", tag)
		},
	}
}

func (fixture *manualReleaseFixture) run(_ context.Context, name string, args ...string) (string, error) {
	call := name + " " + strings.Join(args, " ")
	fixture.calls = append(fixture.calls, call)
	match := call
	if name == "git" {
		match = "git " + strings.Join(args[2:], " ")
	}
	if fixture.fail != "" && fixture.fail != "stamp" && strings.Contains(match, fixture.fail) {
		return "", errors.New("injected command failure")
	}
	if name == "git" {
		return fixture.git(args[2:])
	}
	return fixture.gh(args)
}

func (fixture *manualReleaseFixture) git(args []string) (string, error) {
	switch args[0] {
	case "status", "add":
		return "", nil
	case "rev-parse":
		return fixture.head, nil
	case "tag":
		return args[len(args)-1], nil
	case "diff":
		for _, root := range []string{"cli/dispatcher", "cli/project-runner", "Packages/src"} {
			if fixture.changed[root] && slices.Contains(args, root) {
				return root + "/changed.cs", nil
			}
		}
		return "", nil
	case "log":
		return "- fix: handle UTF-8 paths (abc123)\r\n- chore(main): release native CLI (fff123)\r\n", nil
	case "-c":
		fixture.count++
		fixture.head = fmt.Sprintf("commit%d", fixture.count)
		return "", nil
	case "push":
		fixture.remote = fixture.head
		return "", nil
	default:
		return "", fmt.Errorf("unexpected git arguments %v", args)
	}
}

func (fixture *manualReleaseFixture) gh(args []string) (string, error) {
	switch args[0] {
	case "api":
		if strings.HasSuffix(args[1], "/main") {
			return fixture.remote, nil
		}
		parts := strings.Split(args[1], "/")
		return fixture.tags[parts[len(parts)-1]], nil
	case "release":
		return "false", nil
	case "workflow":
		id := strings.TrimPrefix(args[len(args)-1], "release-id=")
		if args[len(args)-1] == "dry-run=false" {
			id = strings.TrimPrefix(args[len(args)-3], "release-id=")
		}
		fixture.current = manualReleaseWorkflowRun{HeadSHA: fixture.head, DisplayTitle: id, Status: "completed", Conclusion: "success"}
		for _, component := range manualReleaseComponentDefinitions() {
			if args[2] == component.Workflow {
				version, err := readManualReleaseVersion(fixture.root, component.VersionFile, component.VersionField)
				if err != nil {
					return "", err
				}
				fixture.tags[component.TagPrefix+version] = fixture.head
			}
		}
		return "", nil
	case "run":
		data, err := json.Marshal([]manualReleaseWorkflowRun{{DisplayTitle: "old-run", HeadSHA: "old", Conclusion: "success", Status: "completed"}, fixture.current})
		return string(data), err
	default:
		return "", fmt.Errorf("unexpected gh arguments %v", args)
	}
}

// TestManualReleaseRepairsPinAfterNativePublish verifies a package-only retry stamps an already published dispatcher without republishing native components.
func TestManualReleaseRepairsPinAfterNativePublish(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	if err := updateManualReleaseJSON(fixture.root, unityPackageCliPinFile, "dispatcherReleaseTag", "dispatcher-v3.14.1"); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := runManualRelease(context.Background(), &output, fixture.options(), fixture.deps()); err != nil {
		t.Fatal(err)
	}
	tag, err := readManualReleaseVersion(fixture.root, unityPackageCliPinFile, "dispatcherReleaseTag")
	if err != nil || tag != "dispatcher-v3.14.2" {
		t.Fatalf("recovery pin = %s, %v", tag, err)
	}
	if fixture.count != 1 || len(fixture.tags) != 1 || fixture.tags["v3.22.4"] != "commit1" {
		t.Fatalf("package-only recovery unexpectedly republished native components: %+v", fixture.tags)
	}
}

// TestManualReleasePublishesNativeBeforePinnedPackage verifies source commits, release order, and mirrored forward-compatible pins.
func TestManualReleasePublishesNativeBeforePinnedPackage(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	fixture.changed["cli/dispatcher"] = true
	fixture.changed["cli/project-runner"] = true
	var output bytes.Buffer
	if err := runManualRelease(context.Background(), &output, fixture.options(), fixture.deps()); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"dispatcher-v3.14.3": "commit1", "tloop-project-runner-v3.13.2": "commit1", "v3.22.4": "commit2"}
	for tag, commit := range want {
		if fixture.tags[tag] != commit {
			t.Errorf("%s = %s, want %s", tag, fixture.tags[tag], commit)
		}
	}
	joined := strings.Join(fixture.calls, "\n")
	stamp := strings.Index(joined, "stamp dispatcher-v3.14.3")
	if stamp < strings.Index(joined, "workflow run native-cli-publish.yml") || stamp > strings.Index(joined, "workflow run unity-package-release.yml") {
		t.Fatal("pin stamping did not follow native publishing and precede package publishing")
	}
	pin, err := os.ReadFile(filepath.Join(fixture.root, filepath.FromSlash(unityPackageCliPinFile)))
	if err != nil {
		t.Fatal(err)
	}
	mirror, err := os.ReadFile(filepath.Join(fixture.root, ".tloop", "project-runner-pin.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pin, mirror) || !bytes.Contains(pin, []byte("保留")) || !bytes.Contains(pin, []byte("3.13.2")) {
		t.Fatalf("pin content changed or mirrors differ: %s / %s", pin, mirror)
	}
	changelog, err := os.ReadFile(filepath.Join(fixture.root, "Packages", "src", "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(changelog, []byte("\r")) || !bytes.Contains(changelog, []byte("Existing history.")) || bytes.Contains(changelog, []byte("chore(main): release")) {
		t.Fatalf("changelog history or line ending regression: %s", changelog)
	}
}

// TestManualReleaseDoesNotPublishUnchangedOrDryRunSources verifies previews and no-op plans have no write or dispatch effects.
func TestManualReleaseDoesNotPublishUnchangedOrDryRunSources(t *testing.T) {
	for _, dry := range []bool{false, true} {
		t.Run(fmt.Sprint(dry), func(t *testing.T) {
			fixture := newManualReleaseFixture(t)
			options := fixture.options()
			options.DryRun = dry
			fixture.changed["Packages/src"] = dry
			if err := runManualRelease(context.Background(), &bytes.Buffer{}, options, fixture.deps()); err != nil {
				t.Fatal(err)
			}
			if fixture.count != 0 || len(fixture.tags) != 0 {
				t.Fatal("preview or unchanged source published a release")
			}
			version, err := readManualReleaseVersion(fixture.root, "Packages/src/package.json", "version")
			if err != nil || version != "3.22.3" {
				t.Fatalf("preview changed package version: %s, %v", version, err)
			}
		})
	}
}

// TestManualReleaseStopsOnCommandOrAttestationFailure verifies every external command failure prevents later release stages.
func TestManualReleaseStopsOnCommandOrAttestationFailure(t *testing.T) {
	for _, failure := range []string{"status", "rev-parse", "tag --list", "diff --name-only", "/commits/main", "log --no-merges", " add ", "user.name=", "push origin", "workflow run build-and-test", "workflow run dispatcher-publish", "run list", "release view", "/commits/dispatcher-v", "stamp", "workflow run unity-package"} {
		t.Run(failure, func(t *testing.T) {
			fixture := newManualReleaseFixture(t)
			fixture.changed["cli/dispatcher"] = true
			fixture.fail = failure
			err := runManualRelease(context.Background(), &bytes.Buffer{}, fixture.options(), fixture.deps())
			if err == nil {
				t.Fatalf("failure %q did not stop release", failure)
			}
			if fixture.tags["v3.22.4"] != "" {
				t.Fatal("package was published after a failure")
			}
		})
	}
}

// TestManualReleaseStopsWhenMainMoves verifies a concurrent source push cannot be silently included in a release.
func TestManualReleaseStopsWhenMainMoves(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	fixture.changed["Packages/src"] = true
	fixture.remote = "another-commit"
	err := runManualRelease(context.Background(), &bytes.Buffer{}, fixture.options(), fixture.deps())
	if err == nil || !strings.Contains(err.Error(), "main moved") || fixture.count != 0 {
		t.Fatalf("concurrent update was accepted: %v", err)
	}
}
