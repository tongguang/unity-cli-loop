package dispatcher

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	clierrors "github.com/hatayama/unity-cli-loop/common/errors"
	"github.com/hatayama/unity-cli-loop/common/ui"

	"github.com/hatayama/unity-cli-loop/common/clicore"
	"github.com/hatayama/unity-cli-loop/common/project"
	"github.com/hatayama/unity-cli-loop/common/unityprocess"
)

const (
	launchLockfilePoll       = 100 * time.Millisecond
	launchLockfileTimeout    = 5 * time.Second
	launchProcessExitPoll    = 100 * time.Millisecond
	launchProcessExitTimeout = 20 * time.Second
	launchReadinessTimeout   = 10 * time.Minute
	// Why 2s: the UnityLockfile can appear before System Events knows the process.
	// One retry is enough; more retries would delay a launch that already succeeded.
	launchFreshFocusRetryDelay = 2 * time.Second
	projectVersionFilePath     = "ProjectSettings/ProjectVersion.txt"
	recoveryDirectoryPath      = "Assets/_Recovery"
	launchTempDirectoryName    = unityprocess.TempDirectoryName
	unityLockfileName          = unityprocess.LockfileName
)

type launchOptions struct {
	projectPath    string
	restart        bool
	quit           bool
	deleteRecovery bool
	editorVersion  string
	editorPath     string
	platform       string
	maxDepth       int
}

func tryHandleLaunchRequestWithDeps(
	ctx context.Context,
	args []string,
	startPath string,
	globalProjectPath string,
	stdout io.Writer,
	stderr io.Writer,
	deps launchDeps,
) (bool, int) {
	if len(args) == 0 || args[0] != clicore.LaunchCommandName {
		return false, 0
	}
	if clicore.ContainsHelpRequest(args[1:]) {
		printLaunchHelp(stdout)
		return true, 0
	}

	options, err := parseLaunchOptions(args[1:], globalProjectPath)
	if err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{Command: clicore.LaunchCommandName})
		return true, 1
	}

	exitCode := runLaunchWithDeps(ctx, options, startPath, stdout, stderr, deps)
	return true, exitCode
}

func parseLaunchOptions(args []string, globalProjectPath string) (launchOptions, error) {
	options := launchOptions{
		projectPath: globalProjectPath,
		maxDepth:    defaultProjectSearchDepth,
	}

	for index := 0; index < len(args); index++ {
		nextIndex, err := applyLaunchOption(&options, args, index)
		if err != nil {
			return launchOptions{}, err
		}
		index = nextIndex
	}

	return options, validateEditorSelection(options.editorPath, options.editorVersion)
}

func runLaunchWithDeps(ctx context.Context, options launchOptions, startPath string, stdout io.Writer, stderr io.Writer, deps launchDeps) int {
	writeLaunchProjectSearch(stdout, options, startPath)

	projectRoot, err := resolveLaunchProjectRoot(startPath, options)
	if err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{Command: clicore.LaunchCommandName})
		return 1
	}

	if !deleteLaunchRecoveryIfRequested(options, projectRoot, stderr) {
		return 1
	}

	runningProcess, handled, code := findLaunchRunningProcess(ctx, options, projectRoot, stdout, stderr, deps)
	if handled {
		return code
	}

	var previousStop launchStopMethod
	if runningProcess != nil {
		stopMethod, handled, code := handleExistingLaunchProcess(ctx, options, projectRoot, runningProcess, stdout, stderr, deps)
		if handled {
			return code
		}
		previousStop = stopMethod
	}

	if options.quit {
		return writeLaunchQuitResponse(stdout, stderr, projectRoot, nil, "")
	}

	return startUnityAndWaitForReadiness(ctx, options, projectRoot, runningProcess, previousStop, stdout, stderr, deps)
}

func writeLaunchProjectSearch(stdout io.Writer, options launchOptions, startPath string) {
	if options.projectPath != "" {
		return
	}
	depthInfo := strconv.Itoa(options.maxDepth)
	if options.maxDepth == -1 {
		depthInfo = "unlimited"
	}
	clicore.WriteFormat(stdout, "Searching for Unity project under %s (max-depth: %s)...\n\n", startPath, depthInfo)
}

func deleteLaunchRecoveryIfRequested(options launchOptions, projectRoot string, stderr io.Writer) bool {
	if !options.deleteRecovery {
		return true
	}
	if err := os.RemoveAll(filepath.Join(projectRoot, recoveryDirectoryPath)); err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return false
	}
	return true
}

func findLaunchRunningProcess(
	ctx context.Context,
	options launchOptions,
	projectRoot string,
	stdout io.Writer,
	stderr io.Writer,
	deps launchDeps,
) (*unityprocess.UnityProcess, bool, int) {
	runningProcess, err := deps.findRunningUnityProcess(ctx, projectRoot)
	if err == nil {
		return runningProcess, false, 0
	}
	// Sandboxes can block the process scan (e.g. /bin/ps). A responding project IPC
	// proves Unity is running, so plain launch must not fail on the scan alone.
	// Restart and quit still fail because they need a process id to kill.
	if !options.restart && !options.quit && deps.probeProjectIpcFallback(ctx, projectRoot) == nil {
		return nil, true, writeDetectionFallbackLaunchReadyResponse(stdout, stderr, projectRoot, err)
	}
	clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
	return nil, true, 1
}

func handleExistingLaunchProcess(
	ctx context.Context,
	options launchOptions,
	projectRoot string,
	runningProcess *unityprocess.UnityProcess,
	stdout io.Writer,
	stderr io.Writer,
	deps launchDeps,
) (launchStopMethod, bool, int) {
	if !options.restart && !options.quit {
		if options.editorVersion != "" || options.editorPath != "" {
			clierrors.WriteClassifiedError(stderr, launchEditorSelectionRequiresRestartError(options), clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
			return "", true, 1
		}
		return "", true, waitForExistingLaunchReadiness(ctx, projectRoot, runningProcess.Pid, stdout, stderr, deps)
	}
	stopMethod, err := stopUnityProcess(ctx, projectRoot, runningProcess.Pid, deps)
	if err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return "", true, 1
	}
	if options.quit {
		return stopMethod, true, writeLaunchQuitResponse(stdout, stderr, projectRoot, &runningProcess.Pid, stopMethod)
	}
	return stopMethod, false, 0
}

func launchEditorSelectionRequiresRestartError(options launchOptions) error {
	option, value := "--editor-version", options.editorVersion
	commandValue := value
	if options.editorPath != "" {
		option, value = "--editor-path", options.editorPath
		commandValue = `"` + value + `"`
	}
	return &clierrors.ArgumentError{
		Message:  option + " requires --restart when Unity is already running for this project.",
		Option:   option,
		Received: value,
		Command:  clicore.LaunchCommandName,
		NextActions: []string{
			fmt.Sprintf("Run `tloop launch --restart %s %s` to relaunch with the requested Editor.", option, commandValue),
		},
	}
}

func waitForExistingLaunchReadiness(ctx context.Context, projectRoot string, pid int, stdout io.Writer, stderr io.Writer, deps launchDeps) int {
	logLaunchExistingFocusWithDeps(ctx, projectRoot, pid, deps)
	spinner := ui.NewLaunchSpinner(stdout, stderr)
	defer spinner.Stop()
	writeLaunchReadinessWait(stdout, spinner)
	if err := waitForLaunchReadinessWithDeps(ctx, projectRoot, deps); err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return 1
	}
	spinner.Stop()
	return writeExistingLaunchReadyResponse(stdout, stderr, projectRoot, pid)
}

func startUnityAndWaitForReadiness(
	ctx context.Context,
	options launchOptions,
	projectRoot string,
	runningProcess *unityprocess.UnityProcess,
	previousStop launchStopMethod,
	stdout io.Writer,
	stderr io.Writer,
	deps launchDeps,
) int {
	cleanupResult, err := cleanStaleUnityTemp(ctx, projectRoot, deps)
	if err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return 1
	}
	if cleanupResult.lockfileRemoved {
		writeStaleUnityTempCleanupMessage(stdout, projectRoot, cleanupResult.leftoverError == nil)
	}
	if cleanupResult.leftoverError != nil {
		writeStaleUnityTempLeftoverWarning(stderr, cleanupResult.leftoverError)
	}

	unityVersion, err := resolveLaunchEditorVersion(projectRoot, options)
	if err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return 1
	}

	unityPath, err := resolveLaunchExecutablePath(options, unityVersion, deps)
	if err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return 1
	}

	spinner := ui.NewLaunchSpinner(stdout, stderr)
	defer spinner.Stop()

	clicore.WriteLine(stdout, "Opening Editor...")
	clicore.WriteFormat(stdout, "Project Path: %s\n", projectRoot)
	if unityVersion != "" {
		clicore.WriteFormat(stdout, "Detected Tuanjie version: %s\n", unityVersion)
	}
	clicore.WriteFormat(stdout, "Editor Path: %s\n", unityPath)

	command := newUnityLaunchCommand(unityPath, buildUnityLaunchArgs(projectRoot, options))
	if err := command.Start(); err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return 1
	}
	currentPid := command.Process.Pid
	if err := command.Process.Release(); err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return 1
	}
	if err := deps.waitForUnityStartupMarker(ctx, unityLockfilePath(projectRoot), launchLockfilePoll, launchLockfileTimeout); err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return 1
	}
	logLaunchFreshFocusWithDeps(ctx, projectRoot, currentPid, deps)
	writeLaunchReadinessWait(stdout, spinner)
	if err := waitForLaunchReadinessWithDeps(ctx, projectRoot, deps); err != nil {
		clierrors.WriteClassifiedError(stderr, err, clierrors.ErrorContext{ProjectRoot: projectRoot, Command: clicore.LaunchCommandName})
		return 1
	}
	spinner.Stop()
	var previousPid *int
	if runningProcess != nil {
		previousPid = &runningProcess.Pid
	}
	return writeLaunchedReadyResponse(stdout, stderr, projectRoot, previousPid, previousStop, currentPid)
}

func buildUnityLaunchArgs(projectRoot string, options launchOptions) []string {
	launchArgs := []string{"-projectPath", projectRoot}
	if options.platform != "" {
		launchArgs = append(launchArgs, "-buildTarget", options.platform)
	}
	launchArgs = append(launchArgs, "-ignorecompilererrors")
	return launchArgs
}

func newUnityLaunchCommand(unityPath string, launchArgs []string) *exec.Cmd {
	command := newPlatformUnityLaunchCommand(unityPath, launchArgs)
	command.Env = append(os.Environ(), "MSYS_NO_PATHCONV=1")
	configureDetachedUnityLaunchCommand(command)
	return command
}

// A stale lockfile only proves no Unity process is currently running for this project;
// it cannot distinguish a crash from a normal shutdown that left cleanup incomplete, so
// the message must not assert a crash happened.
func writeStaleUnityTempCleanupMessage(stdout io.Writer, projectRoot string, tempDirectoryFullyDeleted bool) {
	clicore.WriteFormat(stdout, "Stale UnityLockfile found (no active Unity process): %s\n", unityLockfilePath(projectRoot))
	clicore.WriteLine(stdout, "Cleaning Temp directory and continuing launch.")
	if tempDirectoryFullyDeleted {
		clicore.WriteLine(stdout, "Deleted Temp directory.")
	}
	clicore.WriteLine(stdout, "Deleted UnityLockfile.")
	clicore.WriteLine(stdout, "")
}

func waitForUnityProcessExit(ctx context.Context, projectRoot string, pid int, pollInterval time.Duration, timeout time.Duration) error {
	return waitForUnityProcessExitWithDeps(ctx, projectRoot, pid, pollInterval, timeout, defaultLaunchDeps())
}

func waitForUnityProcessExitWithDeps(ctx context.Context, projectRoot string, pid int, pollInterval time.Duration, timeout time.Duration, deps launchDeps) error {
	timeoutContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	timeoutError := func() error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return launchProcessExitTimeoutError{
			projectRoot: projectRoot,
			pid:         pid,
			timeout:     timeout,
		}
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		runningProcess, err := deps.findRunningUnityProcess(timeoutContext, projectRoot)
		if err != nil {
			if timeoutContext.Err() != nil {
				return timeoutError()
			}
			return err
		}
		if runningProcess == nil || runningProcess.Pid != pid {
			return nil
		}

		select {
		case <-timeoutContext.Done():
			return timeoutError()
		case <-ticker.C:
		}
	}
}

func waitForUnityStartupMarkerOrTimeout(ctx context.Context, lockfilePath string, pollInterval time.Duration, timeout time.Duration) error {
	timeoutContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if _, err := os.Stat(lockfilePath); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}

		select {
		case <-timeoutContext.Done():
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return nil
		case <-ticker.C:
		}
	}
}

func unityLockfilePath(projectRoot string) string {
	return unityprocess.LockfilePath(projectRoot)
}

func resolveLaunchProjectRoot(startPath string, options launchOptions) (string, error) {
	if options.projectPath != "" {
		return project.ResolveExplicitProjectRoot(options.projectPath)
	}
	return project.FindUnityProjectRootWithin(startPath, options.maxDepth)
}

func printLaunchHelp(stdout io.Writer) {
	clicore.WriteLine(stdout, "Usage:")
	clicore.WriteLine(stdout, "  tloop launch [options] [project-path]")
	clicore.WriteLine(stdout, "")
	clicore.WriteLine(stdout, "Options:")
	clicore.WriteLine(stdout, "  -r, --restart          Quit an existing Unity process for the project before launching")
	clicore.WriteLine(stdout, "  -q, --quit             Quit an existing Unity process for the project without launching")
	clicore.WriteLine(stdout, "  -d, --delete-recovery  Delete Assets/_Recovery before launch")
	clicore.WriteLine(stdout, "      --editor-version <version>")
	clicore.WriteLine(stdout, "                          Use this Tuanjie Editor version instead of ProjectVersion.txt")
	clicore.WriteLine(stdout, "      --editor-path <path>")
	clicore.WriteLine(stdout, "                          Use this Editor executable instead of automatic discovery")
	clicore.WriteLine(stdout, "  -p, --platform <name>  Pass Unity -buildTarget when launching")
	clicore.WriteLine(stdout, "      --max-depth <n>    Max directory depth when auto-searching for the project (default: 3, -1 = unlimited)")
	clicore.WriteLine(stdout, "")
	clicore.WriteLine(stdout, "Compiler errors are ignored by default during Unity startup.")
	clicore.WriteLine(stdout, "")
	printGlobalOptionsHelp(stdout)
	// launch renders its own help instead of going through printNativeSingleCommandHelp, so the
	// closing skill line has to be printed here as well.
	printSkillGuidanceHelp(clicore.LaunchCommandName, stdout)
}
