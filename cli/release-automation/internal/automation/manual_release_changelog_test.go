package automation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Verifies selective merges describe the adopted change without advertising rejected upstream changes.
func TestManualReleaseChangelogUsesResolvedMergeHistory(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	policy := "cli/dispatcher/policy.go"
	writeManualFixtureFile(t, fixture.root, policy, "original policy\n")
	runGitInRepo(t, fixture.root, "init", "-b", "main")
	runGitInRepo(t, fixture.root, "add", ".")
	runGitInRepo(t, fixture.root, "commit", "-m", "Initial release")
	runGitInRepo(t, fixture.root, "tag", "dispatcher-v3.14.2")
	runGitInRepo(t, fixture.root, "checkout", "-b", "upstream")
	writeManualFixtureFile(t, fixture.root, policy, "rejected policy\n")
	runGitInRepo(t, fixture.root, "commit", "-am", "feat: rejected signature policy")
	writeManualFixtureFile(t, fixture.root, "cli/dispatcher/launch.go", "launch guard\n")
	runGitInRepo(t, fixture.root, "add", ".")
	runGitInRepo(t, fixture.root, "commit", "-m", "fix: upstream launch guard")
	runGitInRepo(t, fixture.root, "checkout", "main")
	runGitInRepo(t, fixture.root, "merge", "--no-commit", "--no-ff", "upstream")
	runGitInRepo(t, fixture.root, "restore", "--source=HEAD", "--staged", "--worktree", "--", policy)
	runGitInRepo(t, fixture.root, "commit", "-m", "Merge upstream launch lock protection")
	writeManualFixtureFile(t, fixture.root, "cli/dispatcher/fix.go", "local fix\n")
	runGitInRepo(t, fixture.root, "add", ".")
	runGitInRepo(t, fixture.root, "commit", "-m", "fix: local repair")

	component := manualReleaseComponentDefinitions()[0]
	component.Previous = "3.14.2"
	component.Version = "3.14.3"
	if err := writeManualReleaseChangelog(context.Background(), fixture.options(), component, runCommandOutput); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(fixture.root, component.Root, "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}
	notes := string(data)
	for _, want := range []string{"Merge upstream launch lock protection", "fix: local repair", "Existing history."} {
		if !strings.Contains(notes, want) {
			t.Errorf("changelog missing %q: %s", want, notes)
		}
	}
	if strings.Contains(notes, "rejected signature policy") {
		t.Fatalf("changelog advertises a rejected change: %s", notes)
	}
}
