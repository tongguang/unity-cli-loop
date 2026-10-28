package automation

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	stageUnityPackageCommandName = "stage-unity-package"
	unityPackageManifestEntry    = "package/package.json"
	unityPackageEntryPrefix      = "package/"
	// unityPackageMissingReportLimit keeps the failure message readable when packing drops a
	// whole folder.
	unityPackageMissingReportLimit = 10
	// unityPackageManifestLimit bounds the package.json read so a corrupt entry header
	// cannot make the check allocate an arbitrary amount of memory.
	unityPackageManifestLimit = 1 << 20
)

type unityPackageStageOptions struct {
	Directory string
	Version   string
	Source    string
}

// unityPackageContents is what the staging check learns from one pass over the tarball.
type unityPackageContents struct {
	manifest       []byte
	hasManifest    bool
	regularEntries map[string]bool
}

type unityPackageManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// RunStageUnityPackage checks the single UPM tarball in --dir and renames it
// to the release asset name OpenUPM picks up, printing archive=<path> as a GITHUB_OUTPUT line.
// The tarball must name this package at --version and contain every
// file of the package source in --source, because OpenUPM republishes it without any rebuild.
func RunStageUnityPackage(stdout io.Writer, stderr io.Writer, args []string) int {
	options, err := parseUnityPackageStageOptions(args)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, stageUnityPackageCommandName+":", err)
		return 2
	}
	stagedPath, err := stageUnityPackage(options)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, stageUnityPackageCommandName+":", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "archive=%s\n", stagedPath)
	return 0
}

func parseUnityPackageStageOptions(args []string) (unityPackageStageOptions, error) {
	flags := flag.NewFlagSet(stageUnityPackageCommandName, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	directory := flags.String("dir", "", "directory containing the packed tarball")
	version := flags.String("version", "", "package version the release tag names")
	source := flags.String("source", "", "package source directory the tarball was packed from")
	if err := flags.Parse(args); err != nil {
		return unityPackageStageOptions{}, err
	}
	if *directory == "" {
		return unityPackageStageOptions{}, errors.New("--dir is required")
	}
	if *version == "" {
		return unityPackageStageOptions{}, errors.New("--version is required")
	}
	if *source == "" {
		return unityPackageStageOptions{}, errors.New("--source is required")
	}
	return unityPackageStageOptions{Directory: *directory, Version: *version, Source: *source}, nil
}

func stageUnityPackage(options unityPackageStageOptions) (string, error) {
	archivePath, err := findSingleUnityPackageArchive(options.Directory)
	if err != nil {
		return "", err
	}
	contents, err := readUnityPackageContents(archivePath)
	if err != nil {
		return "", err
	}
	if err := contents.validate(options.Version); err != nil {
		return "", fmt.Errorf("%s: %w", filepath.Base(archivePath), err)
	}
	if err := contents.requireSourceFiles(options.Source); err != nil {
		return "", fmt.Errorf("%s: %w", filepath.Base(archivePath), err)
	}

	stagedPath := filepath.Join(options.Directory, unityPackageAssetName(options.Version))
	if err := os.Rename(archivePath, stagedPath); err != nil {
		return "", fmt.Errorf("rename tarball: %w", err)
	}
	return stagedPath, nil
}

func findSingleUnityPackageArchive(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", fmt.Errorf("read pack output directory: %w", err)
	}
	var archives []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.Type().IsRegular() && (strings.HasSuffix(name, ".tgz") || strings.HasSuffix(name, ".tar.gz")) {
			archives = append(archives, filepath.Join(directory, name))
		}
	}
	if len(archives) != 1 {
		return "", fmt.Errorf("expected exactly one tarball in %s, found %d", directory, len(archives))
	}
	return archives[0], nil
}

func readUnityPackageContents(archivePath string) (unityPackageContents, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return unityPackageContents{}, fmt.Errorf("open tarball: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return unityPackageContents{}, fmt.Errorf("read tarball: %w", err)
	}

	contents := unityPackageContents{regularEntries: map[string]bool{}}
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return contents, nil
		}
		if err != nil {
			return unityPackageContents{}, fmt.Errorf("read tarball: %w", err)
		}
		if err := contents.record(header, tarReader); err != nil {
			return unityPackageContents{}, err
		}
	}
}

func (contents *unityPackageContents) record(header *tar.Header, entryReader io.Reader) error {
	name := header.Name
	if header.Typeflag == tar.TypeReg {
		contents.regularEntries[name] = true
	}
	switch name {
	case unityPackageManifestEntry:
		manifest, err := io.ReadAll(io.LimitReader(entryReader, unityPackageManifestLimit))
		if err != nil {
			return fmt.Errorf("read %s: %w", unityPackageManifestEntry, err)
		}
		contents.manifest = manifest
		contents.hasManifest = true
	}
	return nil
}

func (contents unityPackageContents) validate(version string) error {
	if !contents.hasManifest {
		return fmt.Errorf("%s is missing", unityPackageManifestEntry)
	}
	var manifest unityPackageManifest
	if err := json.Unmarshal(contents.manifest, &manifest); err != nil {
		return fmt.Errorf("parse %s: %w", unityPackageManifestEntry, err)
	}
	if manifest.Name != unityPackageName || manifest.Version != version {
		return fmt.Errorf("%s names %s@%s, want %s@%s", unityPackageManifestEntry, manifest.Name, manifest.Version, unityPackageName, version)
	}
	return nil
}

// requireSourceFiles fails when any regular file of the package source is not a regular file in
// the tarball. A dropped asset .meta file makes Unity import the asset with a new GUID, and a
// dropped tilde-suffixed folder loses the CLI-only skills, so no file may go missing.
func (contents unityPackageContents) requireSourceFiles(sourceDirectory string) error {
	sourceRoot, err := resolvePackageSourceRoot(sourceDirectory)
	if err != nil {
		return err
	}
	var missing []string
	err = filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relativePath, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		// Tar entry names always use "/", while a Windows walk yields backslash-separated paths.
		packedName := unityPackageEntryPrefix + filepath.ToSlash(relativePath)
		if !contents.regularEntries[packedName] {
			missing = append(missing, packedName)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("read package source: %w", err)
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	reported := missing
	if len(reported) > unityPackageMissingReportLimit {
		reported = reported[:unityPackageMissingReportLimit]
	}
	return fmt.Errorf("%d package source files are missing from the tarball: %s", len(missing), strings.Join(reported, ", "))
}

// resolvePackageSourceRoot follows a symlinked root, which filepath.WalkDir would otherwise report
// as a single non-regular entry, and requires the package manifest there, so an empty or wrong
// directory cannot make the completeness check pass with nothing to compare.
func resolvePackageSourceRoot(sourceDirectory string) (string, error) {
	sourceRoot, err := filepath.EvalSymlinks(sourceDirectory)
	if err != nil {
		return "", fmt.Errorf("read package source: %w", err)
	}
	manifest, err := os.Stat(filepath.Join(sourceRoot, "package.json"))
	if err != nil || !manifest.Mode().IsRegular() {
		return "", fmt.Errorf("read package source: %s has no package.json, so it is not a package root", sourceDirectory)
	}
	return sourceRoot, nil
}
