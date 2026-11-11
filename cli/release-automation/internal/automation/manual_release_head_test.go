package automation

import (
	"context"
	"strings"
	"testing"
)

// Verifies a freshly pushed Git ref wins over stale API data while a later remote change still stops release.
func TestManualReleaseHeadReadsPushedRef(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	remote := t.TempDir()
	runGitInRepo(t, remote, "init", "--bare")
	runGitInRepo(t, fixture.root, "init", "-b", "main")
	runGitInRepo(t, fixture.root, "add", ".")
	runGitInRepo(t, fixture.root, "commit", "-m", "Initial release")
	runGitInRepo(t, fixture.root, "remote", "add", "origin", remote)
	runGitInRepo(t, fixture.root, "push", "origin", "HEAD:refs/heads/main")
	head := strings.TrimSpace(runGitInRepoOutput(t, fixture.root, "rev-parse", "HEAD"))
	run := func(ctx context.Context, name string, args ...string) (string, error) {
		if name == "gh" {
			return "stale-api-head", nil
		}
		return runCommandOutput(ctx, name, args...)
	}
	if err := requireManualReleaseHead(context.Background(), fixture.options(), head, run); err != nil {
		t.Fatalf("fresh push rejected: %v", err)
	}
	runGitInRepo(t, fixture.root, "commit", "--allow-empty", "-m", "Another change")
	runGitInRepo(t, fixture.root, "push", "origin", "HEAD:refs/heads/main")
	if err := requireManualReleaseHead(context.Background(), fixture.options(), head, run); err == nil || !strings.Contains(err.Error(), "main moved") {
		t.Fatalf("changed remote was not rejected: %v", err)
	}
}

// Verifies LF and CRLF ref output is accepted, and missing, malformed or unrelated refs fail closed.
func TestManualReleaseHeadValidatesRemoteRefOutput(t *testing.T) {
	for _, test := range []struct {
		name, output string
		valid        bool
	}{
		{"LF", "source\trefs/heads/main\n", true},
		{"CRLF", "source\trefs/heads/main\r\n", true},
		{"missing", "", false},
		{"missing ref", "source\n", false},
		{"other ref", "source\trefs/heads/feature\n", false},
		{"multiple refs", "source\trefs/heads/main\nother\trefs/heads/main\n", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := func(context.Context, string, ...string) (string, error) { return test.output, nil }
			err := requireManualReleaseHead(context.Background(), manualReleaseOptions{}, "source", run)
			if (err == nil) != test.valid {
				t.Fatalf("remote output %q returned %v", test.output, err)
			}
		})
	}
}
