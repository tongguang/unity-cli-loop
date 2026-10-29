package automation

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type manualReleaseRun func(context.Context, string, ...string) (string, error)

type manualReleaseComponent struct {
	Name         string
	Root         string
	VersionFile  string
	VersionField string
	TagPrefix    string
	Workflow     string
	Previous     string
	Version      string
}

type manualReleasePlan struct {
	SourceSHA  string
	Components []manualReleaseComponent
}

func (component manualReleaseComponent) Tag() string {
	return component.TagPrefix + component.Version
}

func (plan manualReleasePlan) nativeComponents() []manualReleaseComponent {
	components := []manualReleaseComponent{}
	for _, component := range plan.Components {
		if component.Name != "package" {
			components = append(components, component)
		}
	}
	return components
}

func manualReleaseComponentDefinitions() []manualReleaseComponent {
	return []manualReleaseComponent{
		{Name: "dispatcher", Root: "cli/dispatcher", VersionFile: "cli/dispatcher/dispatchercontract/dispatcher-contract.json", VersionField: "dispatcherVersion", TagPrefix: "dispatcher-v", Workflow: "dispatcher-publish.yml"},
		{Name: "runner", Root: "cli/project-runner", VersionFile: "cli/common/clicontract/contract.json", VersionField: "projectRunnerVersion", TagPrefix: "tloop-project-runner-v", Workflow: "native-cli-publish.yml"},
		{Name: "package", Root: "Packages/src", VersionFile: "Packages/src/package.json", VersionField: "version", TagPrefix: "v", Workflow: "unity-package-release.yml"},
	}
}

func planManualRelease(ctx context.Context, options manualReleaseOptions, run manualReleaseRun) (manualReleasePlan, error) {
	if !containsManualReleaseChoice(options.Components, "auto", "all", "dispatcher", "runner", "package") {
		return manualReleasePlan{}, fmt.Errorf("invalid components %q", options.Components)
	}
	if !containsManualReleaseChoice(options.Bump, "patch", "minor", "major") {
		return manualReleasePlan{}, fmt.Errorf("invalid bump %q", options.Bump)
	}
	status, err := run(ctx, "git", "-C", options.Root, "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return manualReleasePlan{}, err
	}
	if strings.TrimSpace(status) != "" {
		return manualReleasePlan{}, fmt.Errorf("tracked files have uncommitted changes; commit them before planning a release")
	}
	head, err := run(ctx, "git", "-C", options.Root, "rev-parse", "HEAD")
	if err != nil {
		return manualReleasePlan{}, err
	}
	plan := manualReleasePlan{SourceSHA: strings.TrimSpace(head), Components: []manualReleaseComponent{}}
	nativeChanged := false
	for _, component := range manualReleaseComponentDefinitions() {
		component, selected, err := planManualReleaseComponent(ctx, options, component, run)
		if err != nil {
			return manualReleasePlan{}, err
		}
		if selected || (component.Name == "package" && nativeChanged) {
			plan.Components = append(plan.Components, component)
			nativeChanged = nativeChanged || component.Name != "package"
		}
	}
	return plan, nil
}

func planManualReleaseComponent(ctx context.Context, options manualReleaseOptions, component manualReleaseComponent, run manualReleaseRun) (manualReleaseComponent, bool, error) {
	version, err := readManualReleaseVersion(options.Root, component.VersionFile, component.VersionField)
	if err != nil {
		return component, false, err
	}
	component.Previous = version
	component.Version, err = bumpManualReleaseVersion(version, options.Bump)
	if err != nil {
		return component, false, err
	}
	if options.Components != "auto" {
		return component, options.Components == "all" || options.Components == component.Name, nil
	}
	if component.Name == "package" {
		pending, err := manualReleasePackagePinPending(options.Root)
		if err != nil || pending {
			return component, pending, err
		}
	}
	tag := component.TagPrefix + version
	existing, err := run(ctx, "git", "-C", options.Root, "tag", "--list", tag)
	if err != nil {
		return component, false, err
	}
	// Missing tags include an interrupted release. Advance the version instead
	// of rebuilding an old tag from a different source commit.
	if strings.TrimSpace(existing) == "" {
		return component, true, nil
	}
	args := append([]string{"-C", options.Root, "diff", "--name-only", tag, "HEAD", "--"}, manualReleaseInputs(component)...)
	changed, err := run(ctx, "git", args...)
	return component, strings.TrimSpace(changed) != "", err
}

func manualReleasePackagePinPending(root string) (bool, error) {
	dispatcher := manualReleaseComponentDefinitions()[0]
	version, err := readManualReleaseVersion(root, dispatcher.VersionFile, dispatcher.VersionField)
	if err != nil {
		return false, err
	}
	tag, err := readManualReleaseVersion(root, unityPackageCliPinFile, "dispatcherReleaseTag")
	return tag != dispatcher.TagPrefix+version, err
}

func manualReleaseInputs(component manualReleaseComponent) []string {
	inputs := []string{component.Root}
	if component.Name == "package" {
		return inputs
	}
	inputs = append(inputs, "cli/common", "cli/.go-version", "cli/go.work", "scripts/build-go-cli.sh", "scripts/go-cli-toolchain.sh", ".github/workflows/"+component.Workflow)
	if component.Name == "dispatcher" {
		return append(inputs, "scripts/install.sh", "scripts/install.ps1", "scripts/package-dispatcher.sh")
	}
	return append(inputs, "scripts/package-go-cli.sh")
}

func bumpManualReleaseVersion(version, bump string) (string, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("expected a stable major.minor.patch version, got %q", version)
	}
	numbers := [3]uint64{}
	for i, part := range parts {
		number, err := strconv.ParseUint(part, 10, 32)
		if err != nil || strconv.FormatUint(number, 10) != part {
			return "", fmt.Errorf("invalid stable version %q", version)
		}
		numbers[i] = number
	}
	switch bump {
	case "major":
		numbers = [3]uint64{numbers[0] + 1, 0, 0}
	case "minor":
		numbers = [3]uint64{numbers[0], numbers[1] + 1, 0}
	case "patch":
		numbers[2]++
	default:
		return "", fmt.Errorf("invalid bump %q", bump)
	}
	return fmt.Sprintf("%d.%d.%d", numbers[0], numbers[1], numbers[2]), nil
}

func containsManualReleaseChoice(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}
