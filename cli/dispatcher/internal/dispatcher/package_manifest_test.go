package dispatcher

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// Verifies a bare manifest gains only the Git dependency, without adding scoped registries.
func TestMergePackageManifestAddsOnlyGitDependency(t *testing.T) {
	result, err := mergePackageManifest([]byte(barePackageManifest()), packageGitURL)
	if err != nil || !result.Changed || !result.DependencyAdded {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	assertManifestHasDependency(t, result.Content, dispatcherUnityPackageName, packageGitURL)
	if _, ok := parseOrderedJSONObject(t, result.Content).values["scopedRegistries"]; ok {
		t.Fatalf("unexpected scoped registry: %s", result.Content)
	}
}

// Verifies foreign scoped registries keep all their values when a Git dependency is added.
func TestMergePackageManifestKeepsExistingForeignRegistry(t *testing.T) {
	input := []byte(`{"dependencies":{},"scopedRegistries":[{"name":"Other","url":"https://example.invalid/registry","scopes":["com.other.package"]}]}`)
	result, err := mergePackageManifest(input, packageGitURL)
	if err != nil {
		t.Fatal(err)
	}
	assertManifestRegistriesUnchanged(t, input, result.Content)
}

// Verifies an existing OpenUPM registry is not extended with the unpublished fork's scope.
func TestMergePackageManifestPreservesExistingOpenUPMRegistry(t *testing.T) {
	input := []byte(`{"dependencies":{},"scopedRegistries":[{"name":"OpenUPM","url":"https://package.openupm.com","scopes":["com.other.package"]}]}`)
	result, err := mergePackageManifest(input, packageGitURL)
	if err != nil {
		t.Fatal(err)
	}
	assertManifestRegistriesUnchanged(t, input, result.Content)
}

// Verifies a matching Git dependency leaves the entire manifest byte-identical.
func TestMergePackageManifestNoChangeWhenAlreadyInstalled(t *testing.T) {
	input := []byte(installedPackageManifest(packageGitURL + "#v1.2.3"))
	result, err := mergePackageManifest(input, packageGitURL+"#v1.2.3")
	if err != nil || result.Changed || !bytes.Equal(result.Content, input) {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

// Verifies an existing dependency version is updated in place without reordering keys.
func TestMergePackageManifestUpdatesDependencyVersionInPlace(t *testing.T) {
	input := []byte(`{
  "dependencies": {
    "com.unity.modules.ai": "1.0.0",
    "io.github.tongguang.tloop": "1.2.2",
    "com.unity.modules.animation": "1.0.0"
  }
}
`)

	result, err := mergePackageManifest(input, packageGitURL+"#v1.2.3")
	if err != nil {
		t.Fatalf("mergePackageManifest failed: %v", err)
	}
	if !result.Changed {
		t.Fatal("expected Changed=true")
	}
	if result.PreviousVersion != "1.2.2" {
		t.Fatalf("PreviousVersion mismatch: %q", result.PreviousVersion)
	}
	if result.DependencyAdded {
		t.Fatal("expected DependencyAdded=false for version update")
	}

	deps := decodeOrderedObjectField(t, result.Content, "dependencies")
	keys := deps.keys
	expectedKeys := []string{"com.unity.modules.ai", "io.github.tongguang.tloop", "com.unity.modules.animation"}
	if len(keys) != len(expectedKeys) {
		t.Fatalf("dependency key count mismatch: %#v", keys)
	}
	for index, expected := range expectedKeys {
		if keys[index] != expected {
			t.Fatalf("dependency key order mismatch at %d: %#v", index, keys)
		}
	}
	assertManifestHasDependency(t, result.Content, dispatcherUnityPackageName, packageGitURL+"#v1.2.3")
}

// Verifies non-alphabetical top-level key order is preserved across a rewrite.
func TestMergePackageManifestPreservesTopLevelKeyOrder(t *testing.T) {
	input := []byte(`{
  "testables": [
    "com.unity.testtools.codecoverage"
  ],
  "dependencies": {
    "com.unity.modules.ai": "1.0.0"
  }
}
`)

	result, err := mergePackageManifest(input, packageGitURL+"#v1.2.3")
	if err != nil {
		t.Fatalf("mergePackageManifest failed: %v", err)
	}

	top := parseOrderedJSONObject(t, result.Content)
	if len(top.keys) < 2 || top.keys[0] != "testables" {
		t.Fatalf("top-level key order not preserved: %#v", top.keys)
	}
	if top.keys[1] != "dependencies" {
		t.Fatalf("dependencies should remain second: %#v", top.keys)
	}
}

// Verifies CRLF input is rewritten with CRLF line endings (Windows regression).
func TestMergePackageManifestPreservesCRLFLineEndings(t *testing.T) {
	input := []byte("{\r\n  \"dependencies\": {\r\n    \"com.unity.modules.ai\": \"1.0.0\"\r\n  }\r\n}\r\n")

	result, err := mergePackageManifest(input, packageGitURL+"#v1.2.3")
	if err != nil {
		t.Fatalf("mergePackageManifest failed: %v", err)
	}
	if !bytes.Contains(result.Content, []byte("\r\n")) {
		t.Fatalf("expected CRLF output:\n%q", result.Content)
	}
	if bytes.Contains(bytes.ReplaceAll(result.Content, []byte("\r\n"), nil), []byte("\n")) {
		t.Fatalf("found bare LF in CRLF output:\n%q", result.Content)
	}
}

// Verifies trailing-newline presence or absence matches the input.
func TestMergePackageManifestPreservesMissingTrailingNewline(t *testing.T) {
	withoutNewline := []byte(`{
  "dependencies": {
    "com.unity.modules.ai": "1.0.0"
  }
}`)
	withNewline := append(append([]byte{}, withoutNewline...), '\n')

	resultWithout, err := mergePackageManifest(withoutNewline, packageGitURL+"#v1.2.3")
	if err != nil {
		t.Fatalf("merge without newline failed: %v", err)
	}
	if bytes.HasSuffix(resultWithout.Content, []byte("\n")) {
		t.Fatalf("unexpected trailing newline:\n%q", resultWithout.Content)
	}

	resultWith, err := mergePackageManifest(withNewline, packageGitURL+"#v1.2.3")
	if err != nil {
		t.Fatalf("merge with newline failed: %v", err)
	}
	if !bytes.HasSuffix(resultWith.Content, []byte("\n")) {
		t.Fatalf("expected trailing newline:\n%q", resultWith.Content)
	}
}

// Verifies malformed JSON is rejected.
func TestMergePackageManifestRejectsMalformedJSON(t *testing.T) {
	_, err := mergePackageManifest([]byte(`{not json`), packageGitURL+"#v1.2.3")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

// Verifies trailing garbage after a complete JSON object is rejected.
func TestMergePackageManifestRejectsTrailingGarbage(t *testing.T) {
	_, err := mergePackageManifest([]byte("{\n  \"dependencies\": {}\n}\n trailing"), packageGitURL+"#v1.2.3")
	if err == nil {
		t.Fatal("expected error for trailing garbage after JSON object")
	}
}

func assertManifestHasDependency(t *testing.T, content []byte, name string, version string) {
	t.Helper()
	deps := decodeOrderedObjectField(t, content, "dependencies")
	raw, ok := deps.values[name]
	if !ok {
		t.Fatalf("dependency %s missing from:\n%s", name, content)
	}
	var decoded string
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("dependency value decode failed: %v", err)
	}
	if decoded != version {
		t.Fatalf("dependency version mismatch: got %q want %q", decoded, version)
	}
}

func assertManifestRegistriesUnchanged(t *testing.T, before, after []byte) {
	t.Helper()
	decode := func(content []byte) any {
		t.Helper()
		raw := parseOrderedJSONObject(t, content).values["scopedRegistries"]
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	if !reflect.DeepEqual(decode(before), decode(after)) {
		t.Fatalf("registries changed: before=%s after=%s", before, after)
	}
}

func decodeOrderedObjectField(t *testing.T, content []byte, field string) orderedJSONObject {
	t.Helper()
	top := parseOrderedJSONObject(t, content)
	raw, ok := top.values[field]
	if !ok {
		t.Fatalf("field %s missing from:\n%s", field, content)
	}
	object, err := parseOrderedJSONObjectBytes(raw)
	if err != nil {
		t.Fatalf("parse field %s failed: %v", field, err)
	}
	return object
}

func parseOrderedJSONObject(t *testing.T, content []byte) orderedJSONObject {
	t.Helper()
	object, err := parseOrderedJSONObjectBytes(content)
	if err != nil {
		t.Fatalf("parseOrderedJSONObjectBytes failed: %v", err)
	}
	return object
}

// Verifies malformed owned dependencies are reported rather than overwritten.
func TestMergePackageManifestRejectsMalformedSections(t *testing.T) {
	cases := []struct{ manifest, prefix string }{
		{`{"dependencies":[]}`, "dependencies: "},
		{`{"dependencies":{"io.github.tongguang.tloop":1}}`, "dependency io.github.tongguang.tloop: "},
	}
	for _, tc := range cases {
		result, err := mergePackageManifest([]byte(tc.manifest), packageGitURL)
		if err == nil || !strings.HasPrefix(err.Error(), tc.prefix) || result.Changed {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	}
}

func TestMergePackageManifestAddsMissingDependenciesSection(t *testing.T) {
	// Verifies a manifest without dependencies gains a dependencies object holding only the package.
	result, err := mergePackageManifest([]byte("{}\n"), packageGitURL+"#v1.2.3")

	if err != nil || !result.DependencyAdded || !result.Changed {
		t.Fatalf("unexpected result: %+v err=%v", result, err)
	}
	var manifest struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(result.Content, &manifest); err != nil {
		t.Fatalf("emitted manifest is not JSON: %v\n%s", err, result.Content)
	}
	if len(manifest.Dependencies) != 1 || manifest.Dependencies[dispatcherUnityPackageName] != packageGitURL+"#v1.2.3" {
		t.Fatalf("unexpected dependencies: %v", manifest.Dependencies)
	}
}

func TestMergePackageManifestInsertsDependencyInAlphabeticalOrder(t *testing.T) {
	// Verifies the package is inserted before the first dependency that sorts after it, not appended.
	input := []byte(`{
  "dependencies": {
    "com.unity.modules.ai": "1.0.0",
    "org.example.tool": "2.0.0"
  }
}
`)

	result, err := mergePackageManifest(input, packageGitURL+"#v1.2.3")
	if err != nil {
		t.Fatalf("mergePackageManifest failed: %v", err)
	}
	content := string(result.Content)
	packageIndex := strings.Index(content, dispatcherUnityPackageName)
	if packageIndex < strings.Index(content, "com.unity.modules.ai") || packageIndex > strings.Index(content, "org.example.tool") {
		t.Fatalf("package is not between its neighbors:\n%s", content)
	}
}
