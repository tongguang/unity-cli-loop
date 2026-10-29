package architecture

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// TestManualReleaseWorkflowTopology keeps one CI and one explicit release entry with the three trusted publishers.
func TestManualReleaseWorkflowTopology(t *testing.T) {
	root := findRepositoryRoot(t)
	files := workflowFilePaths(t, root)
	names := []string{}
	for _, filename := range files {
		names = append(names, filepath.Base(filename))
		content, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		var workflow map[string]any
		if err := yaml.Unmarshal(content, &workflow); err != nil {
			t.Fatalf("%s: %v", filename, err)
		}
		events, ok := workflow["on"].(map[string]any)
		if !ok {
			t.Fatalf("%s has no event map", filename)
		}
		if filepath.Base(filename) != "build-and-test.yml" && len(events) != 1 {
			t.Fatalf("%s must publish only on workflow_dispatch", filename)
		}
		if _, exists := events["workflow_dispatch"]; !exists {
			t.Fatalf("%s has no explicit dispatch entry", filename)
		}
		for _, retired := range []string{"create-github-app-token@", "gh pr merge", "gh pr create", "environment: cli-release", "workflow_run:", "HOMEBREW_TAP_TOKEN", "WINGET_PKGS_TOKEN"} {
			if strings.Contains(string(content), retired) {
				t.Errorf("%s still uses retired release machinery %q", filename, retired)
			}
		}
	}
	slices.Sort(names)
	expected := []string{"build-and-test.yml", "dispatcher-publish.yml", "native-cli-publish.yml", "release.yml", "unity-package-release.yml"}
	if !slices.Equal(names, expected) {
		t.Fatalf("workflow entry points = %v, want %v", names, expected)
	}
}

// TestManualPublishersKeepSourceAndArtifactVerification prevents a trigger cleanup from weakening release provenance.
func TestManualPublishersKeepSourceAndArtifactVerification(t *testing.T) {
	root := findRepositoryRoot(t)
	for _, workflow := range []string{"dispatcher-publish.yml", "native-cli-publish.yml"} {
		content, err := os.ReadFile(filepath.Join(root, ".github", "workflows", workflow))
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range []string{"inputs.expected-sha", "inputs.release-id", "github.sha", "TARGET_SHA", "GITHUB_SHA", "actions/attest@", "release-assets.manifest", "--signer-workflow", "--source-digest", "--draft=false"} {
			if !strings.Contains(string(content), required) {
				t.Errorf("%s lost required publisher verification %q", workflow, required)
			}
		}
	}
}
