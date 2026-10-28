package dispatcher

import (
	"io"
	"os"

	"github.com/hatayama/unity-cli-loop/common/clicore"
	"golang.org/x/term"
)

// tryHandleDispatcherHumanVersionRequest answers a plain `-v` / `--version` locally when stdout is a
// terminal, printing the dispatcher version plus one sentence naming the project runner that serves
// this directory.
// Why only on a terminal: the Unity package's CLI detection runs `tloop --version` and `tloop -v` with
// redirected stdout and reads the whole stdout as the version, so the non-terminal contract must stay
// byte-identical.
func tryHandleDispatcherHumanVersionRequest(
	remainingArgs []string,
	projectPath string,
	stdout io.Writer,
	deps dispatcherRunDeps,
) (bool, int) {
	if !clicore.IsVersionRequest(remainingArgs) {
		return false, 0
	}
	if !dispatcherStdoutIsTerminal(stdout, deps) {
		return false, 0
	}
	writeDispatcherVersionOutput(stdout, false)
	clicore.WriteLine(stdout, dispatcherVersionContextSentence(projectPath, remainingArgs))
	return true, 0
}

// dispatcherStdoutIsTerminal reports whether human-readable output is appropriate for stdout.
// Why the nil check: dependency literals built without the probe must behave as "not a terminal".
func dispatcherStdoutIsTerminal(stdout io.Writer, deps dispatcherRunDeps) bool {
	if deps.stdoutIsTerminal == nil {
		return false
	}
	return deps.stdoutIsTerminal(stdout)
}

// isTerminalWriter reports whether the writer is an OS file attached to a terminal.
func isTerminalWriter(writer io.Writer) bool {
	file, isFile := writer.(*os.File)
	if !isFile {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

// dispatcherVersionContextSentence explains which project runner serves the resolved directory.
func dispatcherVersionContextSentence(projectPath string, remainingArgs []string) string {
	projectRoot, err := resolveDispatcherHumanVersionProjectRoot(projectPath, remainingArgs)
	if err != nil {
		return "No Unity project detected here; run inside a Unity project to see which project runner serves it."
	}
	pin, pinErr := loadDispatcherPin(projectRoot)
	if pinErr != nil {
		return "This Unity project has no readable project runner pin yet; " +
			"open it in Unity once to create " + dispatcherProjectPinRelativePath + "."
	}
	return "This Unity project pins tloop project runner " + pin.ProjectRunnerVersion + "."
}

func resolveDispatcherHumanVersionProjectRoot(projectPath string, remainingArgs []string) (string, error) {
	startPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return resolveDispatcherProjectRoot(startPath, projectPath, remainingArgs)
}
