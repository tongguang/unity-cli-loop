package automation

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestManualPackageFailureStopsPublication verifies failed pin writes, package checks and publication verification never advance to a later stage.
func TestManualPackageFailureStopsPublication(t *testing.T) {
	for _, problem := range []string{"main", "dispatcher-version", "mirror", "metadata", "commit", "ci", "tag"} {
		t.Run(problem, func(t *testing.T) {
			fixture := newManualReleaseFixture(t)
			options := fixture.options()
			options.Components = "all"
			plan, err := planManualRelease(context.Background(), options, fixture.run)
			if err != nil {
				t.Fatal(err)
			}
			deps := fixture.deps()
			switch problem {
			case "main":
				fixture.remote = "moved"
			case "dispatcher-version":
				makeManualReleasePathDirectory(t, fixture.root, manualReleaseComponentDefinitions()[0].VersionFile)
			case "mirror":
				makeManualReleasePathDirectory(t, fixture.root, ".tloop/project-runner-pin.json")
			case "metadata":
				writeManualFixtureFile(t, fixture.root, "Packages/src/package.json", "{")
			case "commit":
				fixture.fail = "push origin"
			case "ci":
				fixture.fail = "workflow run build-and-test"
			case "tag":
				deps.run = func(ctx context.Context, name string, args ...string) (string, error) {
					if name == "gh" && args[0] == "release" {
						return "true", nil
					}
					return fixture.run(ctx, name, args...)
				}
			}
			if err := publishManualPackage(context.Background(), io.Discard, options, plan, deps, "source"); err == nil {
				t.Fatal("package stage ignored a failure")
			}
		})
	}
}

func makeManualReleasePathDirectory(t *testing.T, root, name string) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(name))
	if err := os.Remove(filename); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filename, 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestManualMetadataPreparationFailure verifies an invalid manifest, pin or changelog prevents committing incomplete version metadata.
func TestManualMetadataPreparationFailure(t *testing.T) {
	for _, problem := range []string{"manifest", "pin", "mirror", "changelog", "tag-query", "no-notes"} {
		t.Run(problem, func(t *testing.T) {
			fixture := newManualReleaseFixture(t)
			component := manualReleaseComponentDefinitions()[1]
			component.Previous = "3.13.1"
			component.Version = "3.13.2"
			run := fixture.run
			switch problem {
			case "manifest":
				writeManualFixtureFile(t, fixture.root, releasePleaseManifestRelativePath, "{")
			case "pin":
				writeManualFixtureFile(t, fixture.root, unityPackageCliPinFile, "{")
			case "mirror":
				makeManualReleasePathDirectory(t, fixture.root, ".tloop/project-runner-pin.json")
			case "changelog":
				makeManualReleasePathDirectory(t, fixture.root, component.Root+"/CHANGELOG.md")
			case "tag-query":
				fixture.fail = "tag --list"
			case "no-notes":
				run = func(context.Context, string, ...string) (string, error) { return "", nil }
			}
			_, err := prepareManualReleaseFiles(context.Background(), fixture.options(), []manualReleaseComponent{component}, run)
			if (err == nil) != (problem == "no-notes") {
				t.Fatalf("metadata preparation returned %v", err)
			}
		})
	}
}

// TestManualReleaseMainGuardAppliesAtEveryWrite verifies branch movement prevents both version commits and workflow dispatch.
func TestManualReleaseMainGuardAppliesAtEveryWrite(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	fixture.remote = "new-main"
	if _, err := commitManualRelease(context.Background(), fixture.options(), "source", nil, "release", fixture.run); err == nil {
		t.Fatal("commit ignored a moved main")
	}
	if err := dispatchManualReleaseWorkflow(context.Background(), io.Discard, fixture.options(), fixture.deps(), "build-and-test.yml", "source", "ci"); err == nil {
		t.Fatal("dispatch ignored a moved main")
	}
}

type manualReleaseRejectedOutput struct{}

func (manualReleaseRejectedOutput) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

// TestManualReleaseRequiresPlanOutput verifies publication cannot proceed when its proposed versions cannot be reported.
func TestManualReleaseRequiresPlanOutput(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	if err := runManualRelease(context.Background(), manualReleaseRejectedOutput{}, fixture.options(), fixture.deps()); err == nil {
		t.Fatal("release continued without recording its plan")
	}
}

// TestManualReleaseCLIRealPreview verifies a committed checkout can preview versions without GitHub access or any repository mutation.
func TestManualReleaseCLIRealPreview(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	ctx := context.Background()
	for _, args := range [][]string{
		{"init", "--quiet"},
		{"add", "."},
		{"-c", "user.name=Release Test", "-c", "user.email=release@example.test", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "--quiet", "-m", "test: seed release metadata"},
	} {
		if _, err := runCommandOutput(ctx, "git", append([]string{"-C", fixture.root}, args...)...); err != nil {
			t.Fatal(err)
		}
	}
	var output bytes.Buffer
	if code := RunRelease(ctx, &output, &output, []string{"--root", fixture.root, "--repo", "test/repo"}); code != 0 {
		t.Fatalf("preview returned %d: %s", code, output.String())
	}
	if !strings.Contains(output.String(), "3.14.3") {
		t.Fatal("preview did not report the next version")
	}
	status, err := runCommandOutput(ctx, "git", "-C", fixture.root, "status", "--porcelain")
	if err != nil || strings.TrimSpace(status) != "" {
		t.Fatalf("preview changed the checkout: %s, %v", status, err)
	}
	if _, err := parseManualReleaseOptions(ctx, nil); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := parseManualReleaseOptions(cancelled, nil); err == nil {
		t.Fatal("cancelled root lookup succeeded")
	}
}

// TestManualReleaseCommitsMatchPackageVerifier verifies generated metadata commits are accepted by the existing UPM release source check.
func TestManualReleaseCommitsMatchPackageVerifier(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	runGitInRepo(t, fixture.root, "init", "-q")
	runGitInRepo(t, fixture.root, "config", "commit.gpgsign", "false")
	runGitInRepo(t, fixture.root, "config", "core.hooksPath", "/dev/null")
	commitReadinessTestRepo(t, fixture.root, "test: seed release metadata")
	options := fixture.options()
	options.Components = "all"
	ctx := context.Background()
	plan, err := planManualRelease(ctx, options, runCommandOutput)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := prepareManualReleaseFiles(ctx, options, plan.nativeComponents(), runCommandOutput); err != nil {
		t.Fatal(err)
	}
	commitReadinessTestRepo(t, fixture.root, "chore(main): release native CLI")
	component := plan.Components[len(plan.Components)-1]
	changelog := component.Root + "/CHANGELOG.md"
	if _, err := findUnityPackageReleaseCommit(ctx, fixture.root, component.Version, changelog); err == nil {
		t.Fatal("native metadata commit was accepted as the later package release")
	}
	if _, err := prepareManualReleaseFiles(ctx, options, []manualReleaseComponent{component}, runCommandOutput); err != nil {
		t.Fatal(err)
	}
	commitReadinessTestRepo(t, fixture.root, "chore(main): release "+component.Version)
	want := strings.TrimSpace(runGitInRepoOutput(t, fixture.root, "rev-parse", "HEAD"))
	got, err := findUnityPackageReleaseCommit(ctx, fixture.root, component.Version, changelog)
	if err != nil || got != want {
		t.Fatalf("package source = %s, %v; want %s", got, err, want)
	}
}
