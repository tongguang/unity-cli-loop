package dispatcher

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/hatayama/unity-cli-loop/common/clitest"
)

// Tests that CLI-only skill discovery excludes skills marked as internal.
func TestCollectSkillDefinitionsIncludesCliOnlyAndSkipsInternal(t *testing.T) {
	projectRoot := t.TempDir()
	writeTestSkill(t, projectRoot, "Packages/src/Editor/CliOnlyTools~/tloop-launch/Skill", `---
name: tloop-launch
---

# tloop launch
`)
	writeTestSkill(t, projectRoot, "Packages/src/Editor/CliOnlyTools~/tloop-internal/Skill", `---
name: tloop-internal
internal: true
---

# internal
`)

	skills, err := collectSkillDefinitions(projectRoot)
	if err != nil {
		t.Fatalf("collectSkillDefinitions failed: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("skill count mismatch: %#v", skills)
	}
	if skills[0].name != "tloop-launch" {
		t.Fatalf("skill name mismatch: %s", skills[0].name)
	}
}

func TestCollectSkillDefinitionsSkipsSingleQuotedInternalSkill(t *testing.T) {
	// Verifies YAML single quotes do not expose internal skills.
	projectRoot := t.TempDir()
	writeTestSkill(t, projectRoot, "Assets/Editor/InternalTool/Skill", `---
name: tloop-internal-tool
internal: 'true'
---

# internal
`)

	skills, err := collectSkillDefinitions(projectRoot)
	if err != nil {
		t.Fatalf("collectSkillDefinitions failed: %v", err)
	}
	if len(skills) != 0 {
		t.Fatalf("internal skill should be hidden: %#v", skills)
	}
}

// Tests that skill discovery includes package, CLI-only, project-local, and cached package skill roots.
func TestCollectSkillDefinitionsIncludesProjectAndPackageRoots(t *testing.T) {
	projectRoot := t.TempDir()
	writeTestSkill(t, projectRoot, "Packages/src/Editor/FirstPartyTools/Compile/Skill", `---
name: tloop-compile
---

# package
`)
	writeTestSkill(t, projectRoot, "Packages/src/Editor/CliOnlyTools~/tloop-launch/Skill", `---
name: tloop-launch
---

# cli-only
`)
	writeTestSkill(t, projectRoot, "Assets/Editor/CustomTools/Hello/Skill", `---
name: tloop-hello
---

# project
`)
	writeTestSkill(t, projectRoot, "Assets/Editor/DirectTool", `---
name: tloop-direct
---

# direct project
`)
	writeTestSkill(t, projectRoot, "Packages/local-package/Editor/LocalTool/Skill", `---
name: tloop-local-package
---

# local package
`)
	manifestLocalPackageRoot := filepath.Join(t.TempDir(), "manifest-local-package")
	writeSkillFile(t, filepath.Join(manifestLocalPackageRoot, "Editor", "ManifestLocalTool", "Skill"), `---
name: tloop-manifest-local-package
---

# manifest local package
`)
	writeManifest(
		t,
		projectRoot,
		`{"dependencies":{"com.example.cached":"1.0.0","com.example.manifest-local":"file:`+filepath.ToSlash(manifestLocalPackageRoot)+`"}}`)
	writeTestSkill(t, projectRoot, "Library/PackageCache/com.example.cached@1.0.0/Editor/CachedTool/Skill", `---
name: tloop-cached-package
---

# cached package
`)

	skills, err := collectSkillDefinitions(projectRoot)
	if err != nil {
		t.Fatalf("collectSkillDefinitions failed: %v", err)
	}

	actualNames := skillNames(skills)
	expectedNames := []string{
		"tloop-cached-package",
		"tloop-compile",
		"tloop-direct",
		"tloop-hello",
		"tloop-launch",
		"tloop-local-package",
		"tloop-manifest-local-package",
	}
	if !reflect.DeepEqual(actualNames, expectedNames) {
		t.Fatalf("skill names mismatch:\nactual:   %#v\nexpected: %#v", actualNames, expectedNames)
	}
}

// Tests that CLI-only and project-local skills win over package-root duplicates.
func TestCollectSkillDefinitionsUsesUnitySideSourcePrecedence(t *testing.T) {
	projectRoot := t.TempDir()
	writeTestSkill(t, projectRoot, "Packages/src/Editor/FirstPartyTools/Compile/Skill", `---
name: tloop-launch
---

# package launch
`)
	writeTestSkill(t, projectRoot, "Packages/src/Editor/FirstPartyTools/ProjectDuplicate/Skill", `---
name: tloop-project
---

# package project
`)
	writeTestSkill(t, projectRoot, "Packages/src/Editor/CliOnlyTools~/tloop-launch/Skill", `---
name: tloop-launch
---

# cli-only launch
`)
	writeTestSkill(t, projectRoot, "Assets/Editor/ProjectDuplicate/Skill", `---
name: tloop-project
---

# asset project
`)

	skills, err := collectSkillDefinitions(projectRoot)
	if err != nil {
		t.Fatalf("collectSkillDefinitions failed: %v", err)
	}

	assertSkillContentContains(t, skills, "tloop-launch", "# cli-only launch")
	assertSkillContentContains(t, skills, "tloop-project", "# asset project")
}

// Tests that direct SKILL.md files without a frontmatter name use their own directory name.
func TestCollectSkillDefinitionsUsesDirectoryNameWhenDirectSkillOmitsName(t *testing.T) {
	projectRoot := t.TempDir()
	writeTestSkill(t, projectRoot, "Assets/Editor/DirectTool", `---
---

# direct project
`)

	skills, err := collectSkillDefinitions(projectRoot)
	if err != nil {
		t.Fatalf("collectSkillDefinitions failed: %v", err)
	}

	actualNames := skillNames(skills)
	expectedNames := []string{"DirectTool"}
	if !reflect.DeepEqual(actualNames, expectedNames) {
		t.Fatalf("skill names mismatch:\nactual:   %#v\nexpected: %#v", actualNames, expectedNames)
	}
}

// Tests that installing and uninstalling a managed skill copies allowed files and removes the skill.
func TestInstallAndUninstallSkillsForTarget(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, `---
name: tloop-sample
---

# sample
`)
	if err := os.MkdirAll(filepath.Join(sourceDir, "references"), 0o755); err != nil {
		t.Fatalf("failed to create references: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "references", "note.md"), []byte("note"), 0o644); err != nil {
		t.Fatalf("failed to write reference: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "SKILL.md.meta"), []byte("meta"), 0o644); err != nil {
		t.Fatalf("failed to write meta: %v", err)
	}

	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte("sample"),
		sourceDirectory: sourceDir,
	}
	target := targetConfigs["claude"]

	result, err := installSkillsForTarget(projectRoot, projectRoot, target, []skillDefinition{skill}, false, true)
	if err != nil {
		t.Fatalf("installSkillsForTarget failed: %v", err)
	}
	if result.installed != 1 || result.updated != 0 || result.skipped != 0 {
		t.Fatalf("install result mismatch: %#v", result)
	}

	installedDir := filepath.Join(projectRoot, ".claude", "skills", "tuanjie-cli-loop", "tloop-sample")
	if _, err := os.Stat(filepath.Join(installedDir, "SKILL.md")); err != nil {
		t.Fatalf("skill file was not installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(installedDir, "SKILL.md.meta")); err == nil {
		t.Fatal("meta file should not be installed")
	}

	removed, notFound, err := uninstallSkillsForTarget(projectRoot, target, []skillDefinition{skill}, false, true)
	if err != nil {
		t.Fatalf("uninstallSkillsForTarget failed: %v", err)
	}
	if removed != 1 || notFound != 0 {
		t.Fatalf("uninstall result mismatch: removed=%d notFound=%d", removed, notFound)
	}
}

// Tests that CRLF-only drift from Windows checkouts does not mark installed skills stale.
func TestSkillStatusIgnoresCRLFLineEndings(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, "---\nname: tloop-sample\n---\n\n# sample\n")
	referencesDir := filepath.Join(sourceDir, "references")
	if err := os.MkdirAll(referencesDir, 0o755); err != nil {
		t.Fatalf("failed to create references: %v", err)
	}
	if err := os.WriteFile(filepath.Join(referencesDir, "note.md"), []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatalf("failed to write source reference: %v", err)
	}

	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte("---\nname: tloop-sample\n---\n\n# sample\n"),
		sourceDirectory: sourceDir,
	}
	installedDir := filepath.Join(projectRoot, ".claude", "skills", managedSkillsDir, "tloop-sample")
	writeRawSkillFile(t, installedDir, "---\r\nname: tloop-sample\r\n---\r\n\r\n# sample\r\n")
	if err := os.MkdirAll(filepath.Join(installedDir, "references"), 0o755); err != nil {
		t.Fatalf("failed to create installed references: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installedDir, "references", "note.md"), []byte("line1\r\nline2\r\n"), 0o644); err != nil {
		t.Fatalf("failed to write installed reference: %v", err)
	}
	installedSkillContent, err := os.ReadFile(filepath.Join(installedDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read installed skill: %v", err)
	}
	if !bytes.Contains(installedSkillContent, []byte("\r\n")) {
		t.Fatal("test setup should keep CRLF line endings in installed SKILL.md")
	}

	status, err := getSkillStatus(filepath.Join(projectRoot, ".claude", "skills"), skill, true)
	if err != nil {
		t.Fatalf("getSkillStatus failed: %v", err)
	}

	if status != "installed" {
		t.Fatalf("status mismatch: %s", status)
	}
}

// Tests that a root-level non-Markdown file placed by another tool does not mark an installed skill outdated.
func TestSkillStatusIgnoresRootLevelNonMarkdownFiles(t *testing.T) {
	baseDir, skill, installedDir := setUpInstalledSampleSkill(t)
	writeSkillDirectoryFile(t, installedDir, "apm.yml", "name: sample\n")

	assertSkillStatus(t, baseDir, skill, "installed")
}

// Tests that an extra root-level Markdown file still marks the skill outdated, matching the Editor-side rule.
func TestSkillStatusDetectsExtraRootLevelMarkdownFile(t *testing.T) {
	baseDir, skill, installedDir := setUpInstalledSampleSkill(t)
	writeSkillDirectoryFile(t, installedDir, "EXTRA.md", "extra\n")

	assertSkillStatus(t, baseDir, skill, "outdated")
}

// Tests that a .gitkeep file inside an installed skill directory is ignored, matching the Editor-side exclusion list.
func TestSkillStatusIgnoresGitkeepFiles(t *testing.T) {
	baseDir, skill, installedDir := setUpInstalledSampleSkill(t)
	writeSkillDirectoryFile(t, installedDir, filepath.Join("references", ".gitkeep"), "")

	assertSkillStatus(t, baseDir, skill, "installed")
}

// setUpInstalledSampleSkill writes a sample skill to a temporary source directory and an
// identical installed copy, so each caller only has to add the one file under test.
func setUpInstalledSampleSkill(t *testing.T) (string, skillDefinition, string) {
	t.Helper()
	const skillFileContent = "---\nname: tloop-sample\n---\n\n# sample\n"
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, skillFileContent)
	writeSkillDirectoryFile(t, sourceDir, filepath.Join("references", "note.md"), "note\n")

	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte(skillFileContent),
		sourceDirectory: sourceDir,
	}
	baseDir := filepath.Join(projectRoot, ".claude", "skills")
	installedDir := getPreferredSkillDir(baseDir, skill.name, true)
	writeRawSkillFile(t, installedDir, skillFileContent)
	writeSkillDirectoryFile(t, installedDir, filepath.Join("references", "note.md"), "note\n")
	return baseDir, skill, installedDir
}

// writeSkillDirectoryFile writes one file inside a skill directory, creating parent directories.
func writeSkillDirectoryFile(t *testing.T, skillDir string, relativePath string, content string) {
	t.Helper()
	fullPath := filepath.Join(skillDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", relativePath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", relativePath, err)
	}
}

// assertSkillStatus fails when the target-mode status of the skill differs from the expected one.
func assertSkillStatus(t *testing.T, baseDir string, skill skillDefinition, expectedStatus string) {
	t.Helper()
	status, err := getSkillStatus(baseDir, skill, true)
	if err != nil {
		t.Fatalf("getSkillStatus failed: %v", err)
	}
	if status != expectedStatus {
		t.Fatalf("status mismatch: %s", status)
	}
}

// Tests that status checks surface inaccessible installed skill directories.
func TestSkillStatusReturnsStatErrors(t *testing.T) {
	projectRoot := t.TempDir()
	baseDir := filepath.Join(projectRoot, ".claude", "skills")
	skill := skillDefinition{name: "tloop-sample"}

	_, err := getSkillStatusWithStat(
		baseDir,
		skill,
		true,
		func(string) (os.FileInfo, error) {
			return nil, os.ErrPermission
		},
	)

	if err == nil {
		t.Fatal("expected status check error")
	}
}

// Tests that installing skills writes deterministic LF line endings.
func TestInstallSkillsNormalizesCRLFLineEndings(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, "---\r\nname: tloop-sample\r\n---\r\n\r\n# sample\r\n")
	referencesDir := filepath.Join(sourceDir, "references")
	if err := os.MkdirAll(referencesDir, 0o755); err != nil {
		t.Fatalf("failed to create references: %v", err)
	}
	if err := os.WriteFile(filepath.Join(referencesDir, "note.md"), []byte("line1\r\nline2\r\n"), 0o644); err != nil {
		t.Fatalf("failed to write source reference: %v", err)
	}

	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte("---\r\nname: tloop-sample\r\n---\r\n\r\n# sample\r\n"),
		sourceDirectory: sourceDir,
	}

	result, err := installSkillsForTarget(projectRoot, projectRoot, targetConfigs["claude"], []skillDefinition{skill}, false, true)
	if err != nil {
		t.Fatalf("installSkillsForTarget failed: %v", err)
	}
	if result.installed != 1 || result.updated != 0 || result.skipped != 0 {
		t.Fatalf("install result mismatch: %#v", result)
	}

	installedReferencePath := filepath.Join(
		projectRoot,
		".claude",
		"skills",
		managedSkillsDir,
		"tloop-sample",
		"references",
		"note.md")
	content, err := os.ReadFile(installedReferencePath)
	if err != nil {
		t.Fatalf("failed to read installed reference: %v", err)
	}
	if strings.Contains(string(content), "\r") {
		t.Fatalf("installed reference kept CRLF line endings: %q", string(content))
	}
}

// Tests that PowerShell scripts keep their source encoding while line endings are normalized.
func TestNormalizeSkillFileContentPreservesUTF16PowerShellEncoding(t *testing.T) {
	source := utf16LittleEndianWithBOM("line1\r\nline2\r\n")
	expected := utf16LittleEndianWithBOM("line1\nline2\n")

	actual := normalizeSkillFileContent("install.ps1", source)

	if !bytes.Equal(actual, expected) {
		t.Fatalf("normalized UTF-16 content mismatch:\nactual:   % x\nexpected: % x", actual, expected)
	}
}

// Tests that installing skills removes disabled and deprecated skill directories from all layouts.
func TestInstallSkillsForTargetRemovesDisabledAndDeprecatedSkills(t *testing.T) {
	projectRoot := t.TempDir()
	enabledSourceDir := filepath.Join(projectRoot, "source", "Enabled", "Skill")
	writeSkillFile(t, enabledSourceDir, `---
name: tloop-enabled-skill
---

# enabled
`)
	disabledSourceDir := filepath.Join(projectRoot, "source", "Disabled", "Skill")
	writeSkillFile(t, disabledSourceDir, `---
name: tloop-disabled-skill
---

# disabled
`)
	writeToolSettings(t, projectRoot, `{"disabledTools":["disabled-skill"]}`)

	target := targetConfigs["claude"]
	skillsRoot := filepath.Join(projectRoot, target.projectDir, "skills")
	disabledFlatDir := filepath.Join(skillsRoot, "tloop-disabled-skill")
	disabledGroupedDir := filepath.Join(skillsRoot, managedSkillsDir, "tloop-disabled-skill")
	deprecatedFlatDir := filepath.Join(skillsRoot, "tloop-capture-window")
	deprecatedGroupedDir := filepath.Join(skillsRoot, managedSkillsDir, "tloop-capture-window")
	writeSkillFile(t, disabledFlatDir, "---\nname: tloop-disabled-skill\n---\n")
	writeSkillFile(t, disabledGroupedDir, "---\nname: tloop-disabled-skill\n---\n")
	writeSkillFile(t, deprecatedFlatDir, "---\nname: tloop-capture-window\n---\n")
	writeSkillFile(t, deprecatedGroupedDir, "---\nname: tloop-capture-window\n---\n")

	result, err := installSkillsForTarget(projectRoot, projectRoot, target, []skillDefinition{
		{
			name:            "tloop-enabled-skill",
			content:         []byte("---\nname: tloop-enabled-skill\n---\n\n# enabled\n"),
			sourceDirectory: enabledSourceDir,
		},
		{
			name:            "tloop-disabled-skill",
			toolName:        "disabled-skill",
			content:         []byte("---\nname: tloop-disabled-skill\n---\n\n# disabled\n"),
			sourceDirectory: disabledSourceDir,
		},
	}, false, true)
	if err != nil {
		t.Fatalf("installSkillsForTarget failed: %v", err)
	}
	if result.installed != 1 || result.updated != 0 || result.skipped != 0 {
		t.Fatalf("install result mismatch: %#v", result)
	}

	enabledDir := filepath.Join(skillsRoot, managedSkillsDir, "tloop-enabled-skill")
	for _, missingDir := range []string{disabledFlatDir, disabledGroupedDir, deprecatedFlatDir, deprecatedGroupedDir} {
		if _, err := os.Stat(missingDir); err == nil {
			t.Fatalf("stale skill should be removed: %s", missingDir)
		}
	}
	if _, err := os.Stat(filepath.Join(enabledDir, "SKILL.md")); err != nil {
		t.Fatalf("enabled skill should be installed: %v", err)
	}
}

// Verifies that installing skills removes each retired skill from both layouts: the Record Input
// and Replay Input skills, whose tools were retired, the Cancel skill, whose command the
// tloop-status skill now documents, and the V3 CLI invocation migration skill, which the
// migration wizard no longer installs.
func TestInstallSkillsForTargetRemovesRetiredSkill(t *testing.T) {
	retiredSkillNames := []string{
		"tloop-record-input",
		"tloop-replay-input",
		"tloop-cancel",
		"tloop-v3-cli-invocation-migration",
	}
	for _, skillName := range retiredSkillNames {
		t.Run(skillName, func(t *testing.T) {
			projectRoot := t.TempDir()
			target := targetConfigs["claude"]
			skillsRoot := filepath.Join(projectRoot, target.projectDir, "skills")
			flatDir := filepath.Join(skillsRoot, skillName)
			groupedDir := filepath.Join(skillsRoot, managedSkillsDir, skillName)
			writeSkillFile(t, flatDir, "---\nname: "+skillName+"\n---\n")
			writeSkillFile(t, groupedDir, "---\nname: "+skillName+"\n---\n")

			result, err := installSkillsForTarget(projectRoot, projectRoot, target, []skillDefinition{}, false, true)
			if err != nil {
				t.Fatalf("installSkillsForTarget failed: %v", err)
			}
			if result.deprecatedRemoved != 2 {
				t.Fatalf("deprecated removal count mismatch: %#v", result)
			}
			for _, missingDir := range []string{flatDir, groupedDir} {
				if _, err := os.Stat(missingDir); err == nil {
					t.Fatalf("retired skill should be removed: %s", missingDir)
				}
			}
		})
	}
}

// Tests that grouped installs migrate an existing legacy flat skill instead of duplicating it.
func TestInstallSkillsForTargetMigratesLegacyFlatSkillToGroupedLayout(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	skillContent := `---
name: tloop-sample
---

# sample
`
	writeSkillFile(t, sourceDir, skillContent)

	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte(strings.ReplaceAll(skillContent, "\r\n", "\n")),
		sourceDirectory: sourceDir,
	}
	target := targetConfigs["claude"]
	baseDir, err := getSkillsBaseDir(projectRoot, target, false)
	if err != nil {
		t.Fatalf("getSkillsBaseDir failed: %v", err)
	}
	flatDir := getPreferredSkillDir(baseDir, skill.name, false)
	groupedDir := getPreferredSkillDir(baseDir, skill.name, true)
	writeSkillFile(t, flatDir, skillContent)

	result, err := installSkillsForTarget(projectRoot, projectRoot, target, []skillDefinition{skill}, false, true)
	if err != nil {
		t.Fatalf("installSkillsForTarget failed: %v", err)
	}
	if result.installed != 0 || result.updated != 0 || result.skipped != 1 {
		t.Fatalf("install result mismatch: %#v", result)
	}
	if _, err := os.Stat(flatDir); err == nil {
		t.Fatal("flat skill should be migrated")
	}
	if _, err := os.Stat(filepath.Join(groupedDir, "SKILL.md")); err != nil {
		t.Fatalf("grouped skill should exist: %v", err)
	}
}

// Tests that flat installs remove a grouped copy of the same managed skill.
func TestInstallSkillsForTargetRemovesGroupedSkillWhenInstallingFlatLayout(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, `---
name: tloop-sample
---

# sample
`)

	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte("---\nname: tloop-sample\n---\n\n# sample\n"),
		sourceDirectory: sourceDir,
	}
	target := targetConfigs["claude"]
	baseDir, err := getSkillsBaseDir(projectRoot, target, false)
	if err != nil {
		t.Fatalf("getSkillsBaseDir failed: %v", err)
	}
	groupedDir := getPreferredSkillDir(baseDir, skill.name, true)
	flatDir := getPreferredSkillDir(baseDir, skill.name, false)
	writeSkillFile(t, groupedDir, "# grouped\n")

	result, err := installSkillsForTarget(projectRoot, projectRoot, target, []skillDefinition{skill}, false, false)
	if err != nil {
		t.Fatalf("installSkillsForTarget failed: %v", err)
	}
	if result.installed != 1 || result.updated != 0 || result.skipped != 0 {
		t.Fatalf("install result mismatch: %#v", result)
	}
	if _, err := os.Stat(groupedDir); err == nil {
		t.Fatal("grouped skill should be removed")
	}
	if _, err := os.Stat(filepath.Join(flatDir, "SKILL.md")); err != nil {
		t.Fatalf("flat skill should be installed: %v", err)
	}
}

// Tests that the public install command writes skills to the flat layout by default.
func TestRunSkillsInstallDefaultsToFlatLayout(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, `---
name: tloop-sample
---

# sample
`)

	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte("---\nname: tloop-sample\n---\n\n# sample\n"),
		sourceDirectory: sourceDir,
	}
	target := targetConfigs["claude"]
	options := skillCommandOptions{
		targets: []skillTarget{target},
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := runSkillsInstall(projectRoot, []skillDefinition{skill}, options, stdout, stderr)

	if code != 0 {
		t.Fatalf("install should succeed: code=%d stderr=%s", code, stderr.String())
	}
	baseDir, err := getSkillsBaseDir(projectRoot, target, false)
	if err != nil {
		t.Fatalf("getSkillsBaseDir failed: %v", err)
	}
	flatDir := getPreferredSkillDir(baseDir, skill.name, false)
	groupedDir := getPreferredSkillDir(baseDir, skill.name, true)
	if _, err := os.Stat(filepath.Join(flatDir, "SKILL.md")); err != nil {
		t.Fatalf("flat skill should be installed: %v", err)
	}
	if _, err := os.Stat(groupedDir); err == nil {
		t.Fatal("grouped skill should not be installed by default")
	}
}

// Tests that the public list command reads the flat layout by default.
func TestRunSkillsListDefaultsToFlatLayout(t *testing.T) {
	projectRoot := t.TempDir()
	skill := skillDefinition{
		name:    "tloop-sample",
		content: []byte("---\nname: tloop-sample\n---\n\n# sample\n"),
	}
	target := targetConfigs["claude"]
	baseDir, err := getSkillsBaseDir(projectRoot, target, false)
	if err != nil {
		t.Fatalf("getSkillsBaseDir failed: %v", err)
	}
	writeSkillFile(t, getPreferredSkillDir(baseDir, skill.name, false), string(skill.content))
	options := skillCommandOptions{
		targets: []skillTarget{target},
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := runSkillsList(projectRoot, []skillDefinition{skill}, options, stdout, stderr)

	if code != 0 {
		t.Fatalf("list should succeed: code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "tloop-sample (installed)") {
		t.Fatalf("flat skill should be reported as installed: %s", stdout.String())
	}
}

// Tests that the list command reports a skill whose tool is disabled as disabled rather than as a missing install.
func TestRunSkillsListReportsSkillOfDisabledToolAsDisabled(t *testing.T) {
	projectRoot := t.TempDir()
	writeToolSettings(t, projectRoot, `{"disabledTools":["disabled-skill"]}`)
	skill := skillDefinition{
		name:     "tloop-disabled-skill",
		toolName: "disabled-skill",
		content:  []byte("---\nname: tloop-disabled-skill\n---\n"),
	}
	options := skillCommandOptions{
		targets: []skillTarget{targetConfigs["claude"]},
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := runSkillsList(projectRoot, []skillDefinition{skill}, options, stdout, stderr)

	if code != 0 {
		t.Fatalf("list should succeed: code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "tloop-disabled-skill (disabled)") {
		t.Fatalf("skill of a disabled tool should be reported as disabled: %s", stdout.String())
	}
}

// Tests that the list command still reports an enabled tool's uninstalled skill as not installed.
func TestRunSkillsListReportsSkillOfEnabledToolAsNotInstalled(t *testing.T) {
	projectRoot := t.TempDir()
	writeToolSettings(t, projectRoot, `{"disabledTools":["other-skill"]}`)
	skill := skillDefinition{
		name:     "tloop-enabled-skill",
		toolName: "enabled-skill",
		content:  []byte("---\nname: tloop-enabled-skill\n---\n"),
	}
	options := skillCommandOptions{
		targets: []skillTarget{targetConfigs["claude"]},
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := runSkillsList(projectRoot, []skillDefinition{skill}, options, stdout, stderr)

	if code != 0 {
		t.Fatalf("list should succeed: code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "tloop-enabled-skill (not installed)") {
		t.Fatalf("skill of an enabled tool should stay not installed: %s", stdout.String())
	}
}

// Tests that the public uninstall command removes only the flat layout by default.
func TestRunSkillsUninstallDefaultsToFlatLayout(t *testing.T) {
	projectRoot := t.TempDir()
	skill := skillDefinition{
		name:    "tloop-sample",
		content: []byte("sample"),
	}
	target := targetConfigs["claude"]
	baseDir, err := getSkillsBaseDir(projectRoot, target, false)
	if err != nil {
		t.Fatalf("getSkillsBaseDir failed: %v", err)
	}
	flatDir := getPreferredSkillDir(baseDir, skill.name, false)
	groupedDir := getPreferredSkillDir(baseDir, skill.name, true)
	writeSkillFile(t, flatDir, "# flat\n")
	writeSkillFile(t, groupedDir, "# grouped\n")
	options := skillCommandOptions{
		targets: []skillTarget{target},
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := runSkillsUninstall(projectRoot, []skillDefinition{skill}, options, stdout, stderr)

	if code != 0 {
		t.Fatalf("uninstall should succeed: code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(flatDir); err == nil {
		t.Fatal("flat skill should be removed")
	}
	if _, err := os.Stat(groupedDir); err != nil {
		t.Fatalf("grouped skill should remain: %v", err)
	}
}

// Tests that uninstalling uses only the selected layout and leaves the other layout intact.
func TestUninstallSkillsForTargetUsesSelectedLayoutOnly(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, `---
name: tloop-sample
---

# sample
`)

	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte("sample"),
		sourceDirectory: sourceDir,
	}
	target := targetConfigs["claude"]
	baseDir, err := getSkillsBaseDir(projectRoot, target, false)
	if err != nil {
		t.Fatalf("getSkillsBaseDir failed: %v", err)
	}
	groupedDir := getPreferredSkillDir(baseDir, skill.name, true)
	flatDir := getPreferredSkillDir(baseDir, skill.name, false)
	writeSkillFile(t, groupedDir, "# grouped\n")
	writeSkillFile(t, flatDir, "# flat\n")

	removed, notFound, err := uninstallSkillsForTarget(projectRoot, target, []skillDefinition{skill}, false, true)
	if err != nil {
		t.Fatalf("uninstallSkillsForTarget failed: %v", err)
	}
	if removed != 1 || notFound != 0 {
		t.Fatalf("uninstall result mismatch: removed=%d notFound=%d", removed, notFound)
	}
	if _, err := os.Stat(groupedDir); err == nil {
		t.Fatal("grouped skill should be removed")
	}
	if _, err := os.Stat(flatDir); err != nil {
		t.Fatalf("flat skill should remain: %v", err)
	}
}

// Tests that uninstalling deprecated skills only cleans the selected layout.
func TestUninstallSkillsForTargetRemovesDeprecatedSkillsFromSelectedLayoutOnly(t *testing.T) {
	projectRoot := t.TempDir()
	target := targetConfigs["claude"]
	baseDir, err := getSkillsBaseDir(projectRoot, target, false)
	if err != nil {
		t.Fatalf("getSkillsBaseDir failed: %v", err)
	}
	groupedDeprecatedDir := getPreferredSkillDir(baseDir, "tloop-capture-window", true)
	flatDeprecatedDir := getPreferredSkillDir(baseDir, "tloop-capture-window", false)
	writeSkillFile(t, groupedDeprecatedDir, "# grouped deprecated\n")
	writeSkillFile(t, flatDeprecatedDir, "# flat deprecated\n")

	removed, notFound, err := uninstallSkillsForTarget(projectRoot, target, []skillDefinition{}, false, true)
	if err != nil {
		t.Fatalf("uninstallSkillsForTarget failed: %v", err)
	}
	if removed != 1 || notFound != 0 {
		t.Fatalf("uninstall result mismatch: removed=%d notFound=%d", removed, notFound)
	}
	if _, err := os.Stat(groupedDeprecatedDir); err == nil {
		t.Fatal("grouped deprecated skill should be removed")
	}
	if _, err := os.Stat(flatDeprecatedDir); err != nil {
		t.Fatalf("flat deprecated skill should remain: %v", err)
	}
}

// Tests that global skill paths fail instead of falling back to a relative directory.
func TestGetSkillsBaseDirReturnsHomeLookupErrorForGlobalTargets(t *testing.T) {
	originalUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return "", os.ErrPermission
	}
	t.Cleanup(func() {
		userHomeDir = originalUserHomeDir
	})

	_, err := getSkillsBaseDir(t.TempDir(), targetConfigs["claude"], true)

	if err == nil {
		t.Fatal("expected home lookup error")
	}
}

// Tests that skills option parsing rejects unknown flags.
func TestParseSkillsOptionsRequiresKnownFlags(t *testing.T) {
	_, err := parseSkillsOptions([]string{"--claude", "--bad-target"})
	if err == nil {
		t.Fatal("expected unknown option error")
	}
}

// Tests that repeated target flags are ignored after their first occurrence.
func TestParseSkillsOptionsDeduplicatesTargets(t *testing.T) {
	options, err := parseSkillsOptions([]string{"--claude", "--claude", "--codex"})
	if err != nil {
		t.Fatalf("parseSkillsOptions failed: %v", err)
	}

	actualIDs := []string{}
	for _, target := range options.targets {
		actualIDs = append(actualIDs, target.id)
	}
	expectedIDs := []string{"claude", "codex"}
	if !reflect.DeepEqual(actualIDs, expectedIDs) {
		t.Fatalf("target ids mismatch: %#v", actualIDs)
	}
}

// Tests that allSkillTargetIDs and targetConfigs stay in sync so help output,
// flag parsing, and default selection cannot silently drift apart.
func TestAllSkillTargetIDsMatchesTargetConfigs(t *testing.T) {
	idsFromSlice := append([]string{}, allSkillTargetIDs...)
	sort.Strings(idsFromSlice)

	idsFromMap := []string{}
	for id := range targetConfigs {
		idsFromMap = append(idsFromMap, id)
	}
	sort.Strings(idsFromMap)

	if !reflect.DeepEqual(idsFromSlice, idsFromMap) {
		t.Fatalf("allSkillTargetIDs and targetConfigs must have identical keys: slice=%v map=%v", idsFromSlice, idsFromMap)
	}
}

// Tests that defaultSkillTargetIDs resolves to the exact expected targets in
// display order, so an accidental edit to allSkillTargetIDs or
// nonDefaultSkillTargetIDs cannot silently change which targets install by
// default.
func TestDefaultSkillTargetIDsDerivedFromAllSkillTargetIDs(t *testing.T) {
	expected := []string{"claude", "codex", "agents", "antigravity"}

	if !reflect.DeepEqual(defaultSkillTargetIDs, expected) {
		t.Fatalf("defaultSkillTargetIDs mismatch: got=%v want=%v", defaultSkillTargetIDs, expected)
	}
}

// Tests that skills subcommand help lists every target flag from
// allSkillTargetIDs so no target can be added without appearing in help.
func TestPrintSkillsSubcommandHelpListsAllTargets(t *testing.T) {
	stdout := &bytes.Buffer{}
	printSkillsSubcommandHelp("install", stdout)
	output := stdout.String()
	for _, id := range allSkillTargetIDs {
		flag := "--" + id
		if !strings.Contains(output, flag) {
			t.Fatalf("subcommand help missing %s:\n%s", flag, output)
		}
	}
}

// Tests that target guidance printed on missing-target invocations lists every
// target flag from allSkillTargetIDs.
func TestPrintSkillsTargetGuidanceListsAllTargets(t *testing.T) {
	stdout := &bytes.Buffer{}
	printSkillsTargetGuidance("install", stdout)
	output := stdout.String()
	for _, id := range allSkillTargetIDs {
		flag := "--" + id
		if !strings.Contains(output, flag) {
			t.Fatalf("target guidance missing %s:\n%s", flag, output)
		}
	}
}

// Tests that unknown skills subcommands are rejected before project resolution.
func TestTryHandleSkillsRequestRejectsUnknownSubcommandWithoutProject(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handled, code := tryHandleSkillsRequest(
		[]string{"skills", "unknown"},
		t.TempDir(),
		"",
		stdout,
		stderr,
	)

	if !handled || code != 1 {
		t.Fatalf("unexpected result: handled=%v code=%d", handled, code)
	}
	if !strings.Contains(stderr.String(), "Unknown skills command: unknown") {
		t.Fatalf("stderr mismatch: %s", stderr.String())
	}
	if strings.Contains(stderr.String(), "unity project not found") {
		t.Fatalf("unknown subcommand should not resolve project first: %s", stderr.String())
	}
}

// Tests that install refreshes targets that already hold tloop skills even when
// they were not requested, because stale per-agent copies (e.g. .codex) have
// repeatedly misled agents with retired flags.
func TestRunSkillsInstallAutoRefreshesPreviouslyInstalledTargets(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, `---
name: tloop-sample
---

# new content
`)
	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte("---\nname: tloop-sample\n---\n\n# new content\n"),
		sourceDirectory: sourceDir,
	}

	codexTarget := targetConfigs["codex"]
	codexBaseDir, err := getSkillsBaseDir(projectRoot, codexTarget, false)
	if err != nil {
		t.Fatalf("getSkillsBaseDir failed: %v", err)
	}
	staleDir := getPreferredSkillDir(codexBaseDir, skill.name, false)
	writeSkillFile(t, staleDir, "---\nname: tloop-sample\n---\n\n# stale content\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	options := skillCommandOptions{targets: []skillTarget{targetConfigs["claude"]}}
	code := runSkillsInstall(projectRoot, []skillDefinition{skill}, options, stdout, stderr)

	if code != 0 {
		t.Fatalf("install should succeed: code=%d stderr=%s", code, stderr.String())
	}
	refreshedContent, err := os.ReadFile(filepath.Join(staleDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("codex skill should still exist: %v", err)
	}
	if !strings.Contains(string(refreshedContent), "# new content") {
		t.Fatalf("codex skill should be refreshed: %s", string(refreshedContent))
	}
	if !strings.Contains(stdout.String(), "Codex CLI") {
		t.Fatalf("output should mention the auto-refreshed target: %s", stdout.String())
	}
}

// Tests that install never creates skill directories for targets that were
// neither requested nor previously installed.
func TestRunSkillsInstallLeavesUninstalledTargetsUntouched(t *testing.T) {
	projectRoot := t.TempDir()
	sourceDir := filepath.Join(projectRoot, "source", "Skill")
	writeSkillFile(t, sourceDir, `---
name: tloop-sample
---

# sample
`)
	skill := skillDefinition{
		name:            "tloop-sample",
		content:         []byte("---\nname: tloop-sample\n---\n\n# sample\n"),
		sourceDirectory: sourceDir,
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	options := skillCommandOptions{targets: []skillTarget{targetConfigs["claude"]}}
	code := runSkillsInstall(projectRoot, []skillDefinition{skill}, options, stdout, stderr)

	if code != 0 {
		t.Fatalf("install should succeed: code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".codex")); !os.IsNotExist(err) {
		t.Fatalf(".codex should not be created without a prior install: %v", err)
	}
}

// writeTestSkill seeds a SKILL.md fixture at projectRoot/relativeDir via the
// shared clitest.WriteSkillFile helper, which owns the CRLF normalization.
func writeTestSkill(t *testing.T, projectRoot string, relativeDir string, content string) {
	t.Helper()
	clitest.WriteSkillFile(t, projectRoot, relativeDir, "SKILL.md", content)
}

func writeManifest(t *testing.T, projectRoot string, content string) {
	t.Helper()
	manifestDir := filepath.Join(projectRoot, "Packages")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("failed to create manifest dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(manifestDir, "manifest.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
}

func writeToolSettings(t *testing.T, projectRoot string, content string) {
	t.Helper()
	settingsDir := filepath.Join(projectRoot, ".tloop")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("failed to create settings dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.tools.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write tool settings: %v", err)
	}
}

func writeSkillFile(t *testing.T, skillDir string, content string) {
	t.Helper()
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	normalizedContent := strings.ReplaceAll(content, "\r\n", "\n")
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(normalizedContent), 0o644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}
}

func writeRawSkillFile(t *testing.T, skillDir string, content string) {
	t.Helper()
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write raw skill file: %v", err)
	}
}

func utf16LittleEndianWithBOM(content string) []byte {
	bytes := []byte{0xff, 0xfe}
	for _, char := range content {
		bytes = append(bytes, byte(char), byte(char>>8))
	}
	return bytes
}

func skillNames(skills []skillDefinition) []string {
	names := make([]string, 0, len(skills))
	for _, skill := range skills {
		names = append(names, skill.name)
	}
	sort.Strings(names)
	return names
}

func assertSkillContentContains(t *testing.T, skills []skillDefinition, skillName string, expectedContent string) {
	t.Helper()
	for _, skill := range skills {
		if skill.name != skillName {
			continue
		}
		if !strings.Contains(string(skill.content), expectedContent) {
			t.Fatalf("skill %s content mismatch: %s", skillName, string(skill.content))
		}
		return
	}
	t.Fatalf("skill not found: %s", skillName)
}

// lockSkillsTestDirectory sets POSIX permissions on a directory and restores them so t.TempDir can clean up.
func lockSkillsTestDirectory(t *testing.T, directory string, mode os.FileMode) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory permissions are required for this failure.")
	}
	if os.Geteuid() == 0 {
		t.Skip("root ignores file and directory permissions.")
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("failed to create %s: %v", directory, err)
	}
	if err := os.Chmod(directory, mode); err != nil {
		t.Fatalf("failed to chmod %s: %v", directory, err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(directory, 0o755)
	})
}

func sampleSkillWithSource(t *testing.T) skillDefinition {
	t.Helper()
	sourceDir := filepath.Join(t.TempDir(), "Skill")
	writeSkillFile(t, sourceDir, sampleSkillContent)
	return skillDefinition{name: "tloop-sample", content: []byte(sampleSkillContent), sourceDirectory: sourceDir}
}

func TestInstallSkillsForTargetGlobalIgnoresProjectToolSettings(t *testing.T) {
	// Verifies a global install installs skills of tools disabled in the project, since global installs have no project settings.
	projectRoot := t.TempDir()
	writeToolSettings(t, projectRoot, `{"disabledTools":["sample"]}`)
	homeDir := t.TempDir()
	stubSkillsUserHomeDir(t, homeDir, nil)

	result, err := installSkillsForTarget(projectRoot, projectRoot, targetConfigs["claude"], []skillDefinition{sampleSkillWithSource(t)}, true, false)

	if err != nil || result.installed != 1 {
		t.Fatalf("unexpected result: %+v err=%v", result, err)
	}
	if !fileExists(filepath.Join(homeDir, ".claude", "skills", "tloop-sample", "SKILL.md")) {
		t.Fatal("global install must write under the home directory")
	}
}

func TestInstallSkillsForTargetReportsFailures(t *testing.T) {
	// Verifies each install stage failure is returned instead of reporting a partial install as success.
	cases := []struct {
		name        string
		grouped     bool
		setup       func(t *testing.T, baseDir string, skill *skillDefinition)
		wantMessage string
	}{
		{
			name:        "source missing",
			wantMessage: string(filepath.Separator) + "missing: ",
			setup: func(t *testing.T, baseDir string, skill *skillDefinition) {
				skill.sourceDirectory = filepath.Join(t.TempDir(), "missing")
			},
		},
		{
			name:        "legacy migration blocked",
			grouped:     true,
			wantMessage: "rename ",
			setup: func(t *testing.T, baseDir string, skill *skillDefinition) {
				writeSkillFile(t, filepath.Join(baseDir, "tloop-sample"), sampleSkillContent)
				lockSkillsTestDirectory(t, filepath.Join(baseDir, managedSkillsDir), 0o555)
			},
		},
		{
			name:        "empty managed directory unreadable",
			wantMessage: "open ",
			setup: func(t *testing.T, baseDir string, skill *skillDefinition) {
				lockSkillsTestDirectory(t, filepath.Join(baseDir, managedSkillsDir), 0o300)
			},
		},
		{
			name:        "alternate layout cannot be removed",
			wantMessage: "unlinkat ",
			setup: func(t *testing.T, baseDir string, skill *skillDefinition) {
				alternateDir := filepath.Join(baseDir, managedSkillsDir, "tloop-sample")
				writeSkillFile(t, filepath.Join(alternateDir, "locked"), "x")
				lockSkillsTestDirectory(t, filepath.Join(alternateDir, "locked"), 0o555)
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			projectRoot := t.TempDir()
			baseDir := filepath.Join(projectRoot, ".claude", "skills")
			skill := sampleSkillWithSource(t)
			testCase.setup(t, baseDir, &skill)

			_, err := installSkillsForTarget(projectRoot, projectRoot, targetConfigs["claude"], []skillDefinition{skill}, false, testCase.grouped)
			if err == nil || !strings.Contains(err.Error(), testCase.wantMessage) {
				t.Fatalf("expected an error containing %q, got %v", testCase.wantMessage, err)
			}
		})
	}
}

func TestInstallSkillsForTargetReportsHomeLookupFailure(t *testing.T) {
	// Verifies a global install fails when the home directory cannot be resolved.
	stubSkillsUserHomeDir(t, "", errors.New("home unavailable"))

	_, err := installSkillsForTarget(t.TempDir(), t.TempDir(), targetConfigs["claude"], nil, true, false)

	if err == nil || !strings.Contains(err.Error(), "home unavailable") {
		t.Fatalf("expected the home error, got %v", err)
	}
}

func TestInstallSkillForTargetReportsStatusFailure(t *testing.T) {
	// Verifies a skill whose installed location cannot be inspected is reported rather than overwritten.
	baseDir := t.TempDir()
	blockSkillsDirWithFile(t, baseDir, "tloop-sample")
	result := skillInstallResult{}

	err := installSkillForTarget(baseDir, sampleSkillWithSource(t), nil, false, &result)

	if err == nil || !strings.Contains(err.Error(), "SKILL.md: not a directory") {
		t.Fatalf("expected a status failure, got %v", err)
	}
	assertFileContent(t, filepath.Join(baseDir, "tloop-sample"), "not a directory")
}

func TestUninstallSkillsForTargetReportsFailures(t *testing.T) {
	// Verifies uninstall stops on a home lookup failure and on a skill directory that cannot be removed.
	t.Run("home lookup", func(t *testing.T) {
		stubSkillsUserHomeDir(t, "", errors.New("home unavailable"))
		if _, _, err := uninstallSkillsForTarget(t.TempDir(), targetConfigs["claude"], nil, true, false); err == nil || err.Error() != "home unavailable" {
			t.Fatalf("expected the home error, got %v", err)
		}
	})
	t.Run("removal blocked", func(t *testing.T) {
		projectRoot := t.TempDir()
		lockedDir := filepath.Join(projectRoot, ".claude", "skills", "tloop-sample", "locked")
		writeSkillFile(t, lockedDir, "x")
		lockSkillsTestDirectory(t, lockedDir, 0o555)

		removed, _, err := uninstallSkillsForTarget(projectRoot, targetConfigs["claude"], []skillDefinition{{name: "tloop-sample"}}, false, false)

		if err == nil || removed != 0 || !strings.HasPrefix(err.Error(), "unlinkat "+filepath.Join(lockedDir, "SKILL.md")) {
			t.Fatalf("expected a removal failure, got removed=%d err=%v", removed, err)
		}
	})
}

func TestCollectSkillDefinitionsHandlesEditorFolderLayouts(t *testing.T) {
	// Verifies discovery reads loose SKILL.md files, skips excluded folders, empty Skill folders, and unsafe names.
	projectRoot := t.TempDir()
	writeTestSkill(t, projectRoot, "Assets/Editor/Loose", "---\nname: tloop-loose\n---\n")
	writeDispatcherTestFile(t, filepath.Join(projectRoot, "Assets", "Editor", "Loose", "README.md"), "docs")
	writeTestSkill(t, projectRoot, "Assets/Editor/node_modules/pkg/Skill", "---\nname: tloop-excluded\n---\n")
	writeDispatcherTestFile(t, filepath.Join(projectRoot, "Assets", "Editor", "Empty", "Skill", "notes.md"), "no skill file")
	writeTestSkill(t, projectRoot, "Assets/Editor/Unsafe/Skill", "---\nname: bad/name\n---\n")

	skills, err := collectSkillDefinitions(projectRoot)
	if err != nil {
		t.Fatalf("collectSkillDefinitions failed: %v", err)
	}
	if names := skillNames(skills); !reflect.DeepEqual(names, []string{"tloop-loose"}) {
		t.Fatalf("skill names mismatch: %v", names)
	}
}

func TestCollectSkillDefinitionsReportsUnreadableSources(t *testing.T) {
	// Verifies an unreadable folder or SKILL.md under an Editor folder fails discovery instead of silently dropping skills.
	cases := []struct {
		name        string
		setup       func(t *testing.T, projectRoot string)
		wantMessage string
	}{
		{
			name: "unreadable folder",
			setup: func(t *testing.T, projectRoot string) {
				lockSkillsTestDirectory(t, filepath.Join(projectRoot, "Assets", "Editor", "Locked"), 0o000)
			},
			wantMessage: filepath.Join("Editor", "Locked") + ": permission denied",
		},
		{
			name:        "unreadable SKILL.md in Skill folder",
			wantMessage: filepath.Join("Tool", "Skill", "SKILL.md") + ": permission denied",
			setup: func(t *testing.T, projectRoot string) {
				writeTestSkill(t, projectRoot, "Assets/Editor/Tool/Skill", sampleSkillContent)
				lockSkillsTestFile(t, filepath.Join(projectRoot, "Assets", "Editor", "Tool", "Skill", "SKILL.md"))
			},
		},
		{
			name:        "unreadable loose SKILL.md",
			wantMessage: filepath.Join("Loose", "SKILL.md") + ": permission denied",
			setup: func(t *testing.T, projectRoot string) {
				writeTestSkill(t, projectRoot, "Assets/Editor/Loose", sampleSkillContent)
				lockSkillsTestFile(t, filepath.Join(projectRoot, "Assets", "Editor", "Loose", "SKILL.md"))
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			projectRoot := t.TempDir()
			testCase.setup(t, projectRoot)

			_, err := collectSkillDefinitions(projectRoot)
			if err == nil || !strings.Contains(err.Error(), testCase.wantMessage) {
				t.Fatalf("expected error containing %q, got %v", testCase.wantMessage, err)
			}
		})
	}
}
