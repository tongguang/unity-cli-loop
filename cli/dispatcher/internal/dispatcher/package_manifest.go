package dispatcher

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// packageManifestMergeResult reports how Packages/manifest.json changed.
type packageManifestMergeResult struct {
	Content         []byte
	Changed         bool
	DependencyAdded bool
	PreviousVersion string
}

func mergePackageManifest(content []byte, version string) (packageManifestMergeResult, error) {
	hadCRLF := bytes.Contains(content, []byte("\r\n"))
	hadTrailingNewline := len(content) > 0 && (content[len(content)-1] == '\n')
	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))

	root, err := parseOrderedJSONObjectBytes(normalized)
	if err != nil {
		return packageManifestMergeResult{}, err
	}

	result := packageManifestMergeResult{}
	if err := updatePackageManifestDependencies(&root, version, &result); err != nil {
		return packageManifestMergeResult{}, err
	}

	if !result.Changed {
		result.Content = append([]byte{}, content...)
		return result, nil
	}

	emitted, err := emitOrderedJSONObject(root, 0)
	if err != nil {
		return packageManifestMergeResult{}, err
	}
	if hadCRLF {
		emitted = bytes.ReplaceAll(emitted, []byte("\n"), []byte("\r\n"))
	}
	if hadTrailingNewline {
		if hadCRLF {
			if !bytes.HasSuffix(emitted, []byte("\r\n")) {
				emitted = append(emitted, '\r', '\n')
			}
		} else if !bytes.HasSuffix(emitted, []byte("\n")) {
			emitted = append(emitted, '\n')
		}
	} else {
		emitted = bytes.TrimSuffix(emitted, []byte("\r\n"))
		emitted = bytes.TrimSuffix(emitted, []byte("\n"))
	}
	result.Content = emitted
	return result, nil
}

func updatePackageManifestDependencies(root *orderedJSONObject, version string, result *packageManifestMergeResult) error {
	rawDependencies, ok := root.values["dependencies"]
	if !ok {
		dependencies := orderedJSONObject{
			keys:   []string{},
			values: map[string]json.RawMessage{},
		}
		insertDependencyAlphabetically(&dependencies, dispatcherUnityPackageName, version)
		encoded, err := emitOrderedJSONObject(dependencies, 1)
		if err != nil {
			return err
		}
		root.keys = append(root.keys, "dependencies")
		root.values["dependencies"] = encoded
		result.Changed = true
		result.DependencyAdded = true
		return nil
	}

	dependencies, err := parseOrderedJSONObjectBytes(rawDependencies)
	if err != nil {
		return fmt.Errorf("dependencies: %w", err)
	}

	versionJSON, err := json.Marshal(version)
	if err != nil {
		return err
	}
	existing, exists := dependencies.values[dispatcherUnityPackageName]
	if exists {
		var previous string
		if err := json.Unmarshal(existing, &previous); err != nil {
			return fmt.Errorf("dependency %s: %w", dispatcherUnityPackageName, err)
		}
		if previous == version {
			return nil
		}
		dependencies.values[dispatcherUnityPackageName] = json.RawMessage(versionJSON)
		result.Changed = true
		result.PreviousVersion = previous
	} else {
		insertDependencyAlphabetically(&dependencies, dispatcherUnityPackageName, version)
		result.Changed = true
		result.DependencyAdded = true
	}

	encoded, err := emitOrderedJSONObject(dependencies, 1)
	if err != nil {
		return err
	}
	root.values["dependencies"] = encoded
	return nil
}

func insertDependencyAlphabetically(dependencies *orderedJSONObject, name string, version string) {
	versionJSON, err := json.Marshal(version)
	if err != nil {
		panic(err)
	}
	insertAt := len(dependencies.keys)
	for index, key := range dependencies.keys {
		if key > name {
			insertAt = index
			break
		}
	}
	dependencies.keys = append(dependencies.keys, "")
	copy(dependencies.keys[insertAt+1:], dependencies.keys[insertAt:])
	dependencies.keys[insertAt] = name
	if dependencies.values == nil {
		dependencies.values = map[string]json.RawMessage{}
	}
	dependencies.values[name] = json.RawMessage(versionJSON)
}
