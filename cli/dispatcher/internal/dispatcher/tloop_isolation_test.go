package dispatcher

import (
	"os"
	"path/filepath"
	"testing"
)

// Verifies global install and uninstall leave both original skill layouts untouched.
func TestTloopGlobalSkillsPreserveUloop(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := t.TempDir()
	originalUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) { return homeDir, nil }
	t.Cleanup(func() { userHomeDir = originalUserHomeDir })
	writeTestSkill(t, projectRoot, "Packages/src/Editor/FirstPartyTools/Compile/Skill", "---\nname: tloop-compile\n---\nTuanjie compile\n")
	skills, err := collectSkillDefinitions(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, grouped := range []bool{false, true} {
		baseDir := filepath.Join(homeDir, ".agents", "skills")
		foreignPaths := []string{
			filepath.Join(baseDir, "uloop-compile", "SKILL.md"),
			filepath.Join(baseDir, "uloop-cancel", "SKILL.md"),
			filepath.Join(baseDir, "unity-cli-loop", "uloop-compile", "SKILL.md"),
			filepath.Join(baseDir, "v3-cli-invocation-migration", "SKILL.md"),
		}
		for _, foreignPath := range foreignPaths {
			if err := os.MkdirAll(filepath.Dir(foreignPath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(foreignPath, []byte("original Unity skill\r\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := installSkillsForTarget(projectRoot, projectRoot, targetConfigs["agents"], skills, true, grouped); err != nil {
			t.Fatal(err)
		}
		installedPath := filepath.Join(getPreferredSkillDir(baseDir, "tloop-compile", grouped), "SKILL.md")
		if _, err := os.Stat(installedPath); err != nil {
			t.Fatalf("Tuanjie skill was not installed: %v", err)
		}
		if _, _, err := uninstallSkillsForTarget(projectRoot, targetConfigs["agents"], skills, true, grouped); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(installedPath); !os.IsNotExist(err) {
			t.Fatalf("Tuanjie skill was not removed: %v", err)
		}
		for _, foreignPath := range foreignPaths {
			content, err := os.ReadFile(foreignPath)
			if err != nil || string(content) != "original Unity skill\r\n" {
				t.Fatalf("original skill changed at %s: %q, %v", foreignPath, content, err)
			}
		}
	}
}

// Verifies original Unity skill sources cannot be adopted by the Tuanjie installer.
func TestTloopSkillDiscoveryIgnoresUloopSources(t *testing.T) {
	projectRoot := t.TempDir()
	writeTestSkill(t, projectRoot, "Assets/Editor/UnityTool/Skill", "---\nname: uloop-compile\n---\nUnity\n")
	writeTestSkill(t, projectRoot, "Assets/Editor/TuanjieTool/Skill", "---\nname: tloop-compile\n---\nTuanjie\n")
	skills, err := collectSkillDefinitions(projectRoot)
	if err != nil || len(skills) != 1 || skills[0].name != "tloop-compile" {
		t.Fatalf("unexpected Tuanjie skill sources: %#v, %v", skills, err)
	}
}
