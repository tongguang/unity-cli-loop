package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func readManualReleaseVersion(root, relativePath, field string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		return "", err
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(content, &values); err != nil {
		return "", err
	}
	var version string
	if err := json.Unmarshal(values[field], &version); err != nil {
		return "", fmt.Errorf("read %s in %s: %w", field, relativePath, err)
	}
	return version, nil
}

func prepareManualReleaseFiles(ctx context.Context, options manualReleaseOptions, components []manualReleaseComponent, run manualReleaseRun) ([]string, error) {
	paths := []string{releasePleaseManifestRelativePath}
	for _, component := range components {
		if err := updateManualReleaseJSON(options.Root, component.VersionFile, component.VersionField, component.Version); err != nil {
			return nil, err
		}
		if err := updateManualReleaseJSON(options.Root, releasePleaseManifestRelativePath, component.Root, component.Version); err != nil {
			return nil, err
		}
		if component.Name == "runner" {
			if err := updateManualReleaseJSON(options.Root, unityPackageCliPinFile, "projectRunnerVersion", component.Version); err != nil {
				return nil, err
			}
			if err := mirrorManualReleasePin(options.Root); err != nil {
				return nil, err
			}
			paths = append(paths, unityPackageCliPinFile, ".tloop/project-runner-pin.json")
		}
		if err := writeManualReleaseChangelog(ctx, options, component, run); err != nil {
			return nil, err
		}
		paths = append(paths, component.VersionFile, component.Root+"/CHANGELOG.md")
	}
	return paths, nil
}

func updateManualReleaseJSON(root, relativePath, field, value string) error {
	name := filepath.Join(root, filepath.FromSlash(relativePath))
	content, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(content, &fields); err != nil {
		return err
	}
	encodedValue, _ := json.Marshal(value)
	fields[field] = encodedValue
	// RawMessage keeps unknown fields and their values when the pin grows.
	encoded, _ := json.MarshalIndent(fields, "", "  ")
	return os.WriteFile(name, append(encoded, '\n'), 0o644)
}

func mirrorManualReleasePin(root string) error {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(unityPackageCliPinFile)))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, ".tloop", "project-runner-pin.json"), content, 0o644)
}

func writeManualReleaseChangelog(ctx context.Context, options manualReleaseOptions, component manualReleaseComponent, run manualReleaseRun) error {
	tag := component.TagPrefix + component.Previous
	existing, err := run(ctx, "git", "-C", options.Root, "tag", "--list", tag)
	if err != nil {
		return err
	}
	revision := "HEAD"
	if strings.TrimSpace(existing) != "" {
		revision = tag + "..HEAD"
	}
	args := append([]string{"-C", options.Root, "log", "--no-merges", "--format=- %s (%h)", revision, "--"}, manualReleaseInputs(component)...)
	notes, err := run(ctx, "git", args...)
	if err != nil {
		return err
	}
	lines := []string{}
	for _, line := range strings.Split(strings.ReplaceAll(notes, "\r\n", "\n"), "\n") {
		if line != "" && !strings.HasPrefix(line, "- chore(main): release") {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		lines = append(lines, "- Update release dependencies and packaging.")
	}
	name := filepath.Join(options.Root, filepath.FromSlash(component.Root), "CHANGELOG.md")
	old, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	previous := strings.TrimPrefix(strings.ReplaceAll(string(old), "\r\n", "\n"), "# Changelog\n")
	heading := fmt.Sprintf("# Changelog\n\n## [%s](https://github.com/%s/compare/%s...%s) (%s)\n\n", component.Version, options.Repository, tag, component.Tag(), time.Now().UTC().Format("2006-01-02"))
	return os.WriteFile(name, []byte(heading+strings.Join(lines, "\n")+"\n\n"+strings.TrimLeft(previous, "\n")), 0o644)
}
