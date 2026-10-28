//go:build linux

package dispatcher

import (
	"reflect"
	"testing"
)

// Verifies launch on Linux looks for the Editor under the Unity Hub default location in the home directory.
func TestUnityExecutableCandidatesOnLinuxUsesHubPathUnderHome(t *testing.T) {
	t.Setenv("HOME", "/home/tloop-test")

	candidates := unityExecutableCandidates("6000.0.1f1")

	expected := []string{"/home/tloop-test/Tuanjie/Hub/Editor/6000.0.1f1/Editor/Tuanjie"}
	if !reflect.DeepEqual(candidates, expected) {
		t.Fatalf("expected %v, got %v", expected, candidates)
	}
}
