package automation

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestManualReleaseVersionIncrements verifies explicit stable version increments and rejects ambiguous version input.
func TestManualReleaseVersionIncrements(t *testing.T) {
	for _, item := range []struct{ version, bump, want string }{
		{"3.14.2", "patch", "3.14.3"},
		{"3.14.2", "minor", "3.15.0"},
		{"3.14.2", "major", "4.0.0"},
		{"3.14", "patch", ""},
		{"3.-1.2", "patch", ""},
		{"3.01.2", "patch", ""},
		{"3.14.2-beta.1", "patch", ""},
		{"3.14.2", "other", ""},
	} {
		got, err := bumpManualReleaseVersion(item.version, item.bump)
		if got != item.want || (err != nil) != (item.want == "") {
			t.Errorf("%s %s = %s, %v; want %s", item.version, item.bump, got, err, item.want)
		}
	}
}

// TestManualReleaseForcedSelections verifies native selections always include the later package release and keep independent versions.
func TestManualReleaseForcedSelections(t *testing.T) {
	for _, item := range []struct{ choice, want string }{
		{"dispatcher", "dispatcher,package"}, {"runner", "runner,package"}, {"package", "package"}, {"all", "dispatcher,runner,package"},
	} {
		t.Run(item.choice, func(t *testing.T) {
			fixture := newManualReleaseFixture(t)
			options := fixture.options()
			options.Components = item.choice
			plan, err := planManualRelease(context.Background(), options, fixture.run)
			if err != nil {
				t.Fatal(err)
			}
			names := []string{}
			for _, component := range plan.Components {
				names = append(names, component.Name)
			}
			if strings.Join(names, ",") != item.want {
				t.Fatalf("selected %v, want %s", names, item.want)
			}
			if err := runManualRelease(context.Background(), &bytes.Buffer{}, options, fixture.deps()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestManualReleasePlanRejectsInvalidSources verifies invalid options, dirty files and corrupt metadata fail before publication.
func TestManualReleasePlanRejectsInvalidSources(t *testing.T) {
	for _, problem := range []string{"bump", "components", "dirty", "missing", "json", "field", "version"} {
		t.Run(problem, func(t *testing.T) {
			fixture := newManualReleaseFixture(t)
			options := fixture.options()
			run := fixture.run
			switch problem {
			case "bump":
				options.Bump = "invalid"
			case "components":
				options.Components = "invalid"
			case "dirty":
				run = func(context.Context, string, ...string) (string, error) { return "M tracked-file", nil }
			case "missing":
				options.Root = filepath.Join(fixture.root, "missing")
			case "json":
				writeManualFixtureFile(t, fixture.root, manualReleaseComponentDefinitions()[0].VersionFile, "{")
			case "field":
				writeManualFixtureFile(t, fixture.root, manualReleaseComponentDefinitions()[0].VersionFile, "{}")
			case "version":
				writeManualFixtureFile(t, fixture.root, manualReleaseComponentDefinitions()[0].VersionFile, `{"dispatcherVersion":"invalid"}`)
			}
			if _, err := planManualRelease(context.Background(), options, run); err == nil {
				t.Fatal("invalid source was accepted")
			}
		})
	}
}

// TestManualReleasePendingPinErrors verifies unreadable dependency metadata prevents an automatic package release plan.
func TestManualReleasePendingPinErrors(t *testing.T) {
	for _, relativePath := range []string{unityPackageCliPinFile, manualReleaseComponentDefinitions()[0].VersionFile} {
		fixture := newManualReleaseFixture(t)
		makeManualReleasePathDirectory(t, fixture.root, relativePath)
		if _, err := manualReleasePackagePinPending(fixture.root); err == nil {
			t.Fatalf("accepted unreadable pin metadata %s", relativePath)
		}
		if _, err := planManualRelease(context.Background(), fixture.options(), fixture.run); err == nil {
			t.Fatalf("planned a release with unreadable pin metadata %s", relativePath)
		}
	}
}

// TestManualReleaseMissingTagRollsForward verifies an interrupted version is advanced instead of retagged.
func TestManualReleaseMissingTagRollsForward(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	run := func(ctx context.Context, name string, args ...string) (string, error) {
		if name == "git" && strings.Contains(strings.Join(args, " "), "tag --list") {
			return "", nil
		}
		return fixture.run(ctx, name, args...)
	}
	plan, err := planManualRelease(context.Background(), fixture.options(), run)
	if err != nil || len(plan.Components) != 3 || plan.Components[0].Version != "3.14.3" {
		t.Fatalf("missing tag did not roll forward: %+v, %v", plan, err)
	}
}

// TestManualReleaseWorkflowRunMatching verifies old successes, changed heads and failed runs cannot authorize a release.
func TestManualReleaseWorkflowRunMatching(t *testing.T) {
	for _, item := range []struct {
		status, conclusion, head string
		wantDone, wantError      bool
	}{
		{"queued", "", "expected", false, false},
		{"completed", "success", "expected", true, false},
		{"completed", "failure", "expected", false, true},
		{"completed", "cancelled", "expected", false, true},
		{"completed", "success", "another", false, true},
	} {
		done, err := manualReleaseWorkflowResult(manualReleaseWorkflowRun{Status: item.status, Conclusion: item.conclusion, HeadSHA: item.head}, "expected")
		if done != item.wantDone || (err != nil) != item.wantError {
			t.Fatalf("%+v returned %t, %v", item, done, err)
		}
	}
}

// TestManualReleaseWorkflowWaitIsBounded verifies listing lag, malformed output and cancellation stop without accepting a different run.
func TestManualReleaseWorkflowWaitIsBounded(t *testing.T) {
	for _, problem := range []string{"timeout", "cancel", "json"} {
		t.Run(problem, func(t *testing.T) {
			fixture := newManualReleaseFixture(t)
			deps := fixture.deps()
			deps.run = func(ctx context.Context, name string, args ...string) (string, error) {
				if name == "gh" && args[0] == "run" {
					if problem == "json" {
						return "not JSON", nil
					}
					return `[{"displayTitle":"old-success","headSha":"source","status":"completed","conclusion":"success"}]`, nil
				}
				return fixture.run(ctx, name, args...)
			}
			if problem == "cancel" {
				deps.wait = func(context.Context, time.Duration) error { return context.Canceled }
			}
			err := dispatchManualReleaseWorkflow(context.Background(), io.Discard, fixture.options(), deps, "build-and-test.yml", "source", "ci")
			if err == nil {
				t.Fatal("unmatched workflow result was accepted")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(waitManualRelease(ctx, time.Hour), context.Canceled) {
		t.Fatal("workflow wait ignored cancellation")
	}
	if err := waitManualRelease(context.Background(), time.Nanosecond); err != nil {
		t.Fatal(err)
	}
}

// TestManualReleaseRejectsUnpublishedOrWrongTag verifies successful workflow completion alone cannot substitute for a matching published release.
func TestManualReleaseRejectsUnpublishedOrWrongTag(t *testing.T) {
	for _, reply := range []string{"true", "false"} {
		run := func(context.Context, string, ...string) (string, error) { return reply, nil }
		if err := verifyManualPublishedRelease(context.Background(), manualReleaseOptions{Repository: "test/repo"}, "v1.0.0", "expected", run); err == nil {
			t.Fatal("invalid published state was accepted")
		}
	}
}

// TestManualReleaseFileErrorsStopPreparation verifies unreadable metadata, invalid JSON and unwritable mirrors prevent a release commit.
func TestManualReleaseFileErrorsStopPreparation(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	component := manualReleaseComponentDefinitions()[0]
	component.Version = "3.14.3"
	component.Previous = "3.14.2"
	for _, content := range []string{"{", "{}"} {
		writeManualFixtureFile(t, fixture.root, "bad.json", content)
		if content == "{" && updateManualReleaseJSON(fixture.root, "bad.json", "version", "1.0.0") == nil {
			t.Fatal("invalid JSON was accepted")
		}
	}
	if err := updateManualReleaseJSON(fixture.root, "absent", "version", "1.0.0"); err == nil {
		t.Fatal("missing metadata was accepted")
	}
	if err := mirrorManualReleasePin(filepath.Join(fixture.root, "missing")); err == nil {
		t.Fatal("missing pin was accepted")
	}
	options := fixture.options()
	options.Root = filepath.Join(fixture.root, "absent")
	if _, err := prepareManualReleaseFiles(context.Background(), options, []manualReleaseComponent{component}, fixture.run); err == nil {
		t.Fatal("missing version file was accepted")
	}
	// A non-directory parent reproduces a write failure on every supported OS.
	writeManualFixtureFile(t, fixture.root, "blocked", "file")
	if err := os.WriteFile(filepath.Join(fixture.root, "blocked", "child"), []byte("x"), 0o644); err == nil {
		t.Fatal("write failure fixture unexpectedly succeeded")
	}
}

// TestManualReleaseCLIRequiresExplicitWorkflowForWrites verifies local commands default to previews and reject live publication outside main.
func TestManualReleaseCLIRequiresExplicitWorkflowForWrites(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	t.Setenv("GITHUB_REF", "refs/heads/feature")
	for _, args := range [][]string{{"--unknown"}, {"--root", fixture.root, "extra"}, {"--root", fixture.root, "--dry-run=false"}} {
		if code := RunRelease(context.Background(), io.Discard, io.Discard, args); code != 2 {
			t.Fatalf("invalid CLI input %v returned %d", args, code)
		}
	}
	options, err := parseManualReleaseOptions(context.Background(), []string{"--root", fixture.root})
	if err != nil || !options.DryRun {
		t.Fatalf("local command does not default to preview: %+v, %v", options, err)
	}
	if code := RunRelease(context.Background(), io.Discard, io.Discard, []string{"--root", fixture.root, "--bump", "invalid"}); code != 1 {
		t.Fatalf("invalid plan returned %d", code)
	}
}
