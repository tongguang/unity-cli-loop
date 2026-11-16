package dispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hatayama/unity-cli-loop/common/clicore"
	clierrors "github.com/hatayama/unity-cli-loop/common/errors"
	"github.com/hatayama/unity-cli-loop/common/unityprocess"
	"github.com/hatayama/unity-cli-loop/common/unityprocess/unitydialog"
)

func launchTestDialog() unitydialog.Dialog {
	return unitydialog.Dialog{
		Title:   "Package dependency resolver required",
		Message: "Resolve the package's dependencies now?",
		Buttons: []string{"Resolve", "Later", "Cancel"},
	}
}

// launchBlockedRun is what one launch printed when its wait ended.
type launchBlockedRun struct {
	code     int
	stderr   string
	envelope clierrors.CLIErrorEnvelope
}

func decodeLaunchErrorEnvelope(t *testing.T, code int, stderr string) launchBlockedRun {
	t.Helper()
	var envelope clierrors.CLIErrorEnvelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &envelope); err != nil {
		t.Fatalf("stderr is not a JSON error envelope: %v\n%s", err, stderr)
	}
	return launchBlockedRun{code: code, stderr: stderr, envelope: envelope}
}

// freshV3LaunchDeps launches a fake Tuanjie for a V3 project, whose readiness wait ends with err.
func freshV3LaunchDeps(t *testing.T, err error) launchDeps {
	t.Helper()
	deps := defaultLaunchDeps()
	deps.findRunningUnityProcess = func(context.Context, string) (*unityprocess.UnityProcess, error) { return nil, nil }
	deps.resolveUnityExecutablePath = func(string) (string, error) { return fakeUnityExecutablePath(t), nil }
	deps.waitForUnityStartupMarker = func(context.Context, string, time.Duration, time.Duration) error { return nil }
	deps.focusUnityProcess = func(context.Context, int) error { return nil }
	deps.waitForToolReadiness = func(context.Context, string, time.Duration, clicore.DialogFinder) error { return err }
	deps.canPressDialogs = true
	return deps
}

func runFreshLaunch(t *testing.T, projectRoot string, deps launchDeps) launchBlockedRun {
	t.Helper()
	var stderr bytes.Buffer
	code := runLaunchWithDeps(
		context.Background(),
		launchOptions{projectPath: projectRoot, editorVersion: "2022.3.62t9"},
		projectRoot,
		io.Discard,
		&stderr,
		deps,
	)
	return decodeLaunchErrorEnvelope(t, code, stderr.String())
}

func runExistingLaunch(t *testing.T, projectRoot string, deps launchDeps) launchBlockedRun {
	t.Helper()
	var stderr bytes.Buffer
	code := runLaunchWithDeps(context.Background(), launchOptions{projectPath: projectRoot}, projectRoot, io.Discard, &stderr, deps)
	return decodeLaunchErrorEnvelope(t, code, stderr.String())
}

func blockedReadiness(dialogs ...unitydialog.Dialog) error {
	return clicore.ToolReadinessBlockedByDialogError{Dialogs: dialogs, Cause: errors.New("connection refused")}
}

// assertLaunchBlockedByDialog checks the parts every launch stopped by a dialog shares: the code,
// the exit code, the phase, both retry flags of a plain launch, the dialogs in Details, and the last
// step.
func assertLaunchBlockedByDialog(t *testing.T, run launchBlockedRun, dialogCount int) clierrors.CLIError {
	t.Helper()
	return assertLaunchBlockedByDialogRetry(t, run, dialogCount, true)
}

// assertLaunchBlockedByDialogRetry is assertLaunchBlockedByDialog with the SafeToRetry to expect.
func assertLaunchBlockedByDialogRetry(t *testing.T, run launchBlockedRun, dialogCount int, wantSafeToRetry bool) clierrors.CLIError {
	t.Helper()
	cliError := run.envelope.Error
	if run.code != 1 || cliError.ErrorCode != clierrors.ErrorCodeBlockedByDialog {
		t.Fatalf("exit code = %d, ErrorCode = %q, want 1 and %s\n%s", run.code, cliError.ErrorCode, clierrors.ErrorCodeBlockedByDialog, run.stderr)
	}
	if cliError.Phase != clierrors.ErrorPhaseConnection || !cliError.Retryable || cliError.SafeToRetry != wantSafeToRetry {
		t.Fatalf("Phase = %q, Retryable = %v, SafeToRetry = %v, want connection, true, %v",
			cliError.Phase, cliError.Retryable, cliError.SafeToRetry, wantSafeToRetry)
	}
	dialogs, ok := cliError.Details["Dialogs"].([]any)
	if !ok || len(dialogs) != dialogCount {
		t.Fatalf("Details.Dialogs = %#v, want %d dialogs", cliError.Details["Dialogs"], dialogCount)
	}
	if len(cliError.NextActions) == 0 || cliError.NextActions[len(cliError.NextActions)-1] != launchDialogClosedNextAction {
		t.Fatalf("NextActions = %q, want the last one to be %q", cliError.NextActions, launchDialogClosedNextAction)
	}
	return cliError
}

// assertPressGuidance checks the steps point at press-dialog-button first and, right after it, say
// what to do when the project's package has no such command.
func assertPressGuidance(t *testing.T, cliError clierrors.CLIError) {
	t.Helper()
	if !strings.Contains(cliError.NextActions[0], `tloop press-dialog-button --label "<button>"`) {
		t.Fatalf("NextActions = %q, want the first one to point at press-dialog-button", cliError.NextActions)
	}
	if len(cliError.NextActions) < 2 || cliError.NextActions[1] != launchOldPackageNextAction {
		t.Fatalf("NextActions = %q, want the second one to be %q", cliError.NextActions, launchOldPackageNextAction)
	}
}

// assertManualCloseGuidance checks the steps ask the user to close the dialog and name no tloop
// dialog command.
func assertManualCloseGuidance(t *testing.T, cliError clierrors.CLIError) {
	t.Helper()
	if !strings.Contains(cliError.NextActions[0], "ask them to answer it in Unity") {
		t.Fatalf("NextActions = %q, want the first one to ask the user", cliError.NextActions)
	}
	if slices.Contains(cliError.NextActions, launchOldPackageNextAction) {
		t.Fatalf("NextActions = %q, want no step about an old package without a press to make", cliError.NextActions)
	}
	for _, step := range cliError.NextActions {
		if strings.Contains(step, "press-dialog-button") || strings.Contains(step, "get-dialogs") {
			t.Fatalf("NextActions = %q, want no tloop dialog command", cliError.NextActions)
		}
	}
}

// Verifies a fresh launch whose readiness wait stopped at a dialog with buttons reports
// BLOCKED_BY_DIALOG with the dialog's title, the dialog in Details, and a press as the first step (#9).
func TestRunLaunchReportsADialogThatHoldsAFreshEditor(t *testing.T) {
	projectRoot := createLaunchTestProject(t)

	run := runFreshLaunch(t, projectRoot, freshV3LaunchDeps(t, blockedReadiness(launchTestDialog())))

	cliError := assertLaunchBlockedByDialog(t, run, 1)
	wantMessage := `Unity is running, but a modal dialog "Package dependency resolver required" is open; ` +
		"Unity waits until it is closed. tloop launch stopped waiting; the Editor keeps running."
	if cliError.Message != wantMessage {
		t.Fatalf("Message = %q, want %q", cliError.Message, wantMessage)
	}
	assertPressGuidance(t, cliError)
	if cause := cliError.Details["Cause"]; cause != "connection refused" {
		t.Fatalf("Details.Cause = %#v, want the failure the wait saw last", cause)
	}
}

// Verifies a launch with --restart that stopped at a dialog is not safe to run again as it was, since
// a second restart would quit, and could kill, the Editor the dialog holds; its last step still runs
// plain tloop launch.
func TestRunLaunchWithRestartStoppedAtADialogIsNotSafeToRetry(t *testing.T) {
	restart := func(t *testing.T, projectRoot string, deps launchDeps) launchBlockedRun {
		t.Helper()
		var stderr bytes.Buffer
		code := runLaunchWithDeps(
			context.Background(),
			launchOptions{projectPath: projectRoot, editorVersion: "2022.3.62t9", restart: true},
			projectRoot,
			io.Discard,
			&stderr,
			deps,
		)
		return decodeLaunchErrorEnvelope(t, code, stderr.String())
	}
	t.Run("v3", func(t *testing.T) {
		projectRoot := createLaunchTestProject(t)

		run := restart(t, projectRoot, freshV3LaunchDeps(t, blockedReadiness(launchTestDialog())))

		_ = assertLaunchBlockedByDialogRetry(t, run, 1, false)
	})
}

// Verifies a dialog without buttons is reported with steps that ask the user to close it (#10).
func TestRunLaunchAsksTheUserToCloseADialogWithoutButtons(t *testing.T) {
	projectRoot := createLaunchTestProject(t)
	dialog := unitydialog.Dialog{Title: "Importing", Message: "Please wait", Buttons: []string{}}

	run := runFreshLaunch(t, projectRoot, freshV3LaunchDeps(t, blockedReadiness(dialog)))

	assertManualCloseGuidance(t, assertLaunchBlockedByDialog(t, run, 1))
}

// Verifies a dialog with buttons is reported with steps that ask the user to close it where tloop
// cannot press a button on this operating system.
func TestRunLaunchAsksTheUserToCloseADialogWhereTloopCannotPress(t *testing.T) {
	projectRoot := createLaunchTestProject(t)
	deps := freshV3LaunchDeps(t, blockedReadiness(launchTestDialog()))
	deps.canPressDialogs = false

	run := runFreshLaunch(t, projectRoot, deps)

	assertManualCloseGuidance(t, assertLaunchBlockedByDialog(t, run, 1))
}

// Verifies a dialog without a title is reported without an empty quoted title (#10b).
func TestRunLaunchReportsAnUntitledDialogWithoutQuotes(t *testing.T) {
	projectRoot := createLaunchTestProject(t)
	dialog := unitydialog.Dialog{Title: "", Message: "", Buttons: []string{"OK"}}

	run := runFreshLaunch(t, projectRoot, freshV3LaunchDeps(t, blockedReadiness(dialog)))

	cliError := assertLaunchBlockedByDialog(t, run, 1)
	wantMessage := "Unity is running, but a modal dialog is open; Unity waits until it is closed. " +
		"tloop launch stopped waiting; the Editor keeps running."
	if cliError.Message != wantMessage {
		t.Fatalf("Message = %q, want %q", cliError.Message, wantMessage)
	}
}

// Verifies two dialogs are counted in the message, and the steps answer the frontmost one, as tloop
// status counts them (#10c).
func TestRunLaunchCountsSeveralDialogs(t *testing.T) {
	projectRoot := createLaunchTestProject(t)
	behind := unitydialog.Dialog{Title: "Behind", Buttons: []string{}}

	run := runFreshLaunch(t, projectRoot, freshV3LaunchDeps(t, blockedReadiness(launchTestDialog(), behind)))

	cliError := assertLaunchBlockedByDialog(t, run, 2)
	wantMessage := "Unity is running, but 2 modal dialogs are open; Unity waits until they are closed. " +
		"tloop launch stopped waiting; the Editor keeps running."
	if cliError.Message != wantMessage {
		t.Fatalf("Message = %q, want %q", cliError.Message, wantMessage)
	}
	assertPressGuidance(t, cliError)
}

// Verifies launch for an Editor that is already open reports a dialog the same way (#11).
func TestRunLaunchReportsADialogThatHoldsAnOpenEditor(t *testing.T) {
	projectRoot := createLaunchTestProject(t)
	deps := freshV3LaunchDeps(t, blockedReadiness(launchTestDialog()))
	deps.findRunningUnityProcess = func(context.Context, string) (*unityprocess.UnityProcess, error) {
		return &unityprocess.UnityProcess{Pid: 111}, nil
	}

	run := runExistingLaunch(t, projectRoot, deps)

	assertPressGuidance(t, assertLaunchBlockedByDialog(t, run, 1))
}

// Verifies both V3 readiness waits are handed a dialog look that reads the Editor found by the
// definitive process lookup, so the wait can stop at a dialog at all.
func TestRunLaunchHandsTheReadinessWaitADialogLook(t *testing.T) {
	cases := map[string]func(*testing.T, string, launchDeps) launchBlockedRun{
		"fresh":    runFreshLaunch,
		"existing": runExistingLaunch,
	}
	for name, run := range cases {
		t.Run(name, func(t *testing.T) {
			projectRoot := createLaunchTestProject(t)
			deps := freshV3LaunchDeps(t, nil)
			if name == "existing" {
				deps.findRunningUnityProcess = func(context.Context, string) (*unityprocess.UnityProcess, error) {
					return &unityprocess.UnityProcess{Pid: 111}, nil
				}
			}
			definitiveLookups := 0
			deps.findRunningUnityProcessDefinitively = func(context.Context, string) (*unityprocess.UnityProcess, error) {
				definitiveLookups++
				return &unityprocess.UnityProcess{Pid: 222}, nil
			}
			var lookedAt string
			deps.lookForDialogs = func(ctx context.Context, root string, lookup unitydialog.ProcessLookup) ([]unitydialog.Dialog, error) {
				lookedAt = root
				if _, err := lookup(ctx, root); err != nil {
					return nil, err
				}
				return []unitydialog.Dialog{launchTestDialog()}, nil
			}
			deps.waitForToolReadiness = func(ctx context.Context, _ string, _ time.Duration, findDialogs clicore.DialogFinder) error {
				if findDialogs == nil {
					return errors.New("no dialog look was handed to the wait")
				}
				dialogs, err := findDialogs(ctx)
				if err != nil {
					return err
				}
				return clicore.ToolReadinessBlockedByDialogError{Dialogs: dialogs}
			}

			result := run(t, projectRoot, deps)

			_ = assertLaunchBlockedByDialog(t, result, 1)
			if lookedAt != projectRoot || definitiveLookups != 1 {
				t.Fatalf("look at %q with %d definitive lookups, want %q and 1", lookedAt, definitiveLookups, projectRoot)
			}
		})
	}
}

// Verifies the look launch hands over reports no dialog, without reading Unity's windows, when no
// Unity process runs, when the lookup fails, and on Linux, where it does not look up a process
// either (#8, #8b, #8c); and returns the dialogs the windows show when an Editor runs (#8d).
func TestLaunchDialogLookReadsTheWindowsOfTheRunningEditorOnly(t *testing.T) {
	inspectWith := func(goos string) func(context.Context, string, unitydialog.ProcessLookup) ([]unitydialog.Dialog, error) {
		return func(ctx context.Context, root string, lookup unitydialog.ProcessLookup) ([]unitydialog.Dialog, error) {
			return unitydialog.NewInspector(goos, lookup).Inspect(ctx, root)
		}
	}
	cases := map[string]struct {
		goos        string
		lookup      unitydialog.ProcessLookup
		wantLookups int
	}{
		"no unity process": {goos: "darwin", lookup: func(context.Context, string) (*unityprocess.UnityProcess, error) {
			return nil, nil
		}, wantLookups: 1},
		"lookup failed": {goos: "windows", lookup: func(context.Context, string) (*unityprocess.UnityProcess, error) {
			return nil, errors.New("could not list processes")
		}, wantLookups: 1},
		"linux": {goos: "linux", lookup: func(context.Context, string) (*unityprocess.UnityProcess, error) {
			return &unityprocess.UnityProcess{Pid: 7}, nil
		}, wantLookups: 0},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			lookups := 0
			deps := defaultLaunchDeps()
			deps.lookForDialogs = inspectWith(testCase.goos)
			lookup := func(ctx context.Context, root string) (*unityprocess.UnityProcess, error) {
				lookups++
				return testCase.lookup(ctx, root)
			}

			dialogs, _ := launchDialogFinder("<PROJECT_ROOT>", lookup, deps)(context.Background())

			if len(dialogs) != 0 || lookups != testCase.wantLookups {
				t.Fatalf("dialogs = %#v after %d lookups, want none after %d", dialogs, lookups, testCase.wantLookups)
			}
		})
	}

	t.Run("dialogs shown", func(t *testing.T) {
		deps := defaultLaunchDeps()
		want := []unitydialog.Dialog{launchTestDialog(), {Title: "Behind", Buttons: []string{"OK"}}}
		deps.lookForDialogs = func(context.Context, string, unitydialog.ProcessLookup) ([]unitydialog.Dialog, error) {
			return want, nil
		}

		dialogs, err := launchDialogFinder("<PROJECT_ROOT>", nil, deps)(context.Background())

		if err != nil || len(dialogs) != 2 {
			t.Fatalf("dialogs = %#v, err = %v, want the two dialogs", dialogs, err)
		}
	})
}

// Verifies the launch error keeps the dialog's text out of Error() and is not read as a lost
// connection, even for a title such as "EOF" (#13).
func TestLaunchDialogErrorLeavesTheDialogOutOfItsText(t *testing.T) {
	deps := defaultLaunchDeps()
	deps.waitForToolReadiness = func(context.Context, string, time.Duration, clicore.DialogFinder) error {
		return blockedReadiness(unitydialog.Dialog{Title: "Unexpected EOF", Buttons: []string{"OK"}})
	}

	err := waitForLaunchReadinessWithDeps(context.Background(), t.TempDir(), false, deps)

	if err == nil || strings.Contains(err.Error(), "EOF") || clierrors.IsTransportDisconnectError(err) {
		t.Fatalf("err = %v, want an error that names no dialog and is no lost connection", err)
	}
	cliError := clierrors.ClassifyError(err, clierrors.ErrorContext{Command: clicore.LaunchCommandName})
	if cliError.ErrorCode != clierrors.ErrorCodeBlockedByDialog {
		t.Fatalf("ErrorCode = %q, want %s", cliError.ErrorCode, clierrors.ErrorCodeBlockedByDialog)
	}
}
