package automation

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type manualReleaseOptions struct {
	Root       string
	Repository string
	Bump       string
	Components string
	RunID      string
	DryRun     bool
}

type manualReleaseDeps struct {
	run   func(context.Context, string, ...string) (string, error)
	stamp func(context.Context, string, string) error
	wait  func(context.Context, time.Duration) error
}

// RunRelease previews or executes releases without creating release pull requests.
func RunRelease(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	options, err := parseManualReleaseOptions(ctx, args)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "release:", err)
		return 2
	}
	deps := manualReleaseDeps{run: runCommandOutput, stamp: StampDispatcherPin, wait: waitManualRelease}
	if err := runManualRelease(ctx, stdout, options, deps); err != nil {
		_, _ = fmt.Fprintln(stderr, "release:", err)
		return 1
	}
	return 0
}

func parseManualReleaseOptions(ctx context.Context, args []string) (manualReleaseOptions, error) {
	options := manualReleaseOptions{}
	flags := flag.NewFlagSet("release", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.Root, "root", "", "repository root")
	flags.StringVar(&options.Repository, "repo", os.Getenv("GITHUB_REPOSITORY"), "owner/repository")
	flags.StringVar(&options.Bump, "bump", "patch", "patch, minor, or major")
	flags.StringVar(&options.Components, "components", "auto", "auto, all, dispatcher, runner, or package")
	flags.StringVar(&options.RunID, "run-id", os.Getenv("GITHUB_RUN_ID"), "parent workflow run ID")
	flags.BoolVar(&options.DryRun, "dry-run", true, "preview without changing files or starting workflows")
	if err := flags.Parse(args); err != nil {
		return options, err
	}
	if flags.NArg() != 0 {
		return options, errors.New("unexpected positional arguments")
	}
	if options.Root == "" {
		root, err := gitRepoRoot(ctx)
		if err != nil {
			return options, err
		}
		options.Root = root
	}
	if !options.DryRun && (options.Repository == "" || options.RunID == "" || os.Getenv("GITHUB_REF") != "refs/heads/main") {
		return options, errors.New("publish through release.yml on main; local invocations support --dry-run only")
	}
	return options, nil
}

func runManualRelease(ctx context.Context, out io.Writer, options manualReleaseOptions, deps manualReleaseDeps) error {
	plan, err := planManualRelease(ctx, options, deps.run)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(plan); err != nil {
		return err
	}
	if options.DryRun || len(plan.Components) == 0 {
		return nil
	}
	return publishManualRelease(ctx, out, options, plan, deps)
}

func publishManualRelease(ctx context.Context, out io.Writer, options manualReleaseOptions, plan manualReleasePlan, deps manualReleaseDeps) error {
	if err := requireManualReleaseHead(ctx, options, plan.SourceSHA, deps.run); err != nil {
		return err
	}
	head := plan.SourceSHA
	if natives := plan.nativeComponents(); len(natives) != 0 {
		paths, err := prepareManualReleaseFiles(ctx, options, natives, deps.run)
		if err != nil {
			return err
		}
		head, err = commitManualRelease(ctx, options, head, paths, "chore(main): release native CLI", deps.run)
		if err != nil {
			return err
		}
		if err := dispatchManualReleaseWorkflow(ctx, out, options, deps, "build-and-test.yml", head, "native-ci"); err != nil {
			return err
		}
		for _, component := range natives {
			if err := publishManualNative(ctx, out, options, deps, component, head); err != nil {
				return err
			}
		}
	}
	return publishManualPackage(ctx, out, options, plan, deps, head)
}

func publishManualNative(ctx context.Context, out io.Writer, options manualReleaseOptions, deps manualReleaseDeps, component manualReleaseComponent, head string) error {
	if err := dispatchManualReleaseWorkflow(ctx, out, options, deps, component.Workflow, head, component.Name); err != nil {
		return err
	}
	if err := verifyManualPublishedRelease(ctx, options, component.Tag(), head, deps.run); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, "Published", component.Tag(), "from", head)
	return nil
}

func publishManualPackage(ctx context.Context, out io.Writer, options manualReleaseOptions, plan manualReleasePlan, deps manualReleaseDeps, head string) error {
	if err := requireManualReleaseHead(ctx, options, head, deps.run); err != nil {
		return err
	}
	// A previous run may have published the dispatcher but stopped before stamping its pin.
	dispatcher := manualReleaseComponentDefinitions()[0]
	version, err := readManualReleaseVersion(options.Root, dispatcher.VersionFile, dispatcher.VersionField)
	if err != nil {
		return err
	}
	pin := filepath.Join(options.Root, filepath.FromSlash(unityPackageCliPinFile))
	if err := deps.stamp(ctx, pin, dispatcher.TagPrefix+version); err != nil {
		return err
	}
	if err := mirrorManualReleasePin(options.Root); err != nil {
		return err
	}
	paths := []string{unityPackageCliPinFile, ".tloop/project-runner-pin.json"}
	component := plan.Components[len(plan.Components)-1]
	packagePaths, err := prepareManualReleaseFiles(ctx, options, []manualReleaseComponent{component}, deps.run)
	if err != nil {
		return err
	}
	paths = append(paths, packagePaths...)
	head, err = commitManualRelease(ctx, options, head, paths, "chore(main): release "+component.Version, deps.run)
	if err != nil {
		return err
	}
	if err := dispatchManualReleaseWorkflow(ctx, out, options, deps, "build-and-test.yml", head, "package-ci"); err != nil {
		return err
	}
	if err := dispatchManualReleaseWorkflow(ctx, out, options, deps, component.Workflow, head, "package"); err != nil {
		return err
	}
	if err := verifyManualPublishedRelease(ctx, options, component.Tag(), head, deps.run); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, "Published", component.Tag(), "from", head)
	return nil
}

func requireManualReleaseHead(ctx context.Context, options manualReleaseOptions, expected string, run manualReleaseRun) error {
	actual, err := run(ctx, "gh", "api", "repos/"+options.Repository+"/commits/main", "--jq", ".sha")
	if err != nil {
		return err
	}
	if strings.TrimSpace(actual) != expected {
		return fmt.Errorf("main moved from %s to %s; stop and start a new release from the updated source", expected, strings.TrimSpace(actual))
	}
	return nil
}

func commitManualRelease(ctx context.Context, options manualReleaseOptions, expected string, paths []string, title string, run manualReleaseRun) (string, error) {
	if err := requireManualReleaseHead(ctx, options, expected, run); err != nil {
		return "", err
	}
	commands := [][]string{
		append([]string{"add", "--"}, paths...),
		{"-c", "user.name=github-actions[bot]", "-c", "user.email=41898282+github-actions[bot]@users.noreply.github.com", "-c", "core.hooksPath=/dev/null", "commit", "-m", title},
		{"push", "origin", "HEAD:refs/heads/main"},
	}
	for _, args := range commands {
		if _, err := run(ctx, "git", append([]string{"-C", options.Root}, args...)...); err != nil {
			return "", err
		}
	}
	head, err := run(ctx, "git", "-C", options.Root, "rev-parse", "HEAD")
	return strings.TrimSpace(head), err
}

func verifyManualPublishedRelease(ctx context.Context, options manualReleaseOptions, tag, expected string, run manualReleaseRun) error {
	draft, err := run(ctx, "gh", "release", "view", tag, "--repo", options.Repository, "--json", "isDraft", "--jq", ".isDraft")
	if err != nil {
		return err
	}
	if strings.TrimSpace(draft) != "false" {
		return fmt.Errorf("release %s is not published", tag)
	}
	commit, err := run(ctx, "gh", "api", "repos/"+options.Repository+"/commits/"+tag, "--jq", ".sha")
	if err != nil {
		return err
	}
	if strings.TrimSpace(commit) != expected {
		return fmt.Errorf("release %s points at %s instead of %s", tag, strings.TrimSpace(commit), expected)
	}
	return nil
}

func waitManualRelease(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
