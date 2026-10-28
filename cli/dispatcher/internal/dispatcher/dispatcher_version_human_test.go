package dispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hatayama/unity-cli-loop/common/clicontract"
	"github.com/hatayama/unity-cli-loop/dispatcher/internal/nativepath"
)

func alwaysTerminalStdout() func(io.Writer) bool {
	return func(io.Writer) bool { return true }
}

func TestRunDispatcherPrintsDispatcherVersionAndPinContextWhenStdoutIsTerminal(t *testing.T) {
	// Verifies an interactive --version in a V3 project prints the pinned project runner version instead of forwarding.
	projectRoot := createDispatcherUnityProject(t)
	writeDispatcherProjectPin(t, projectRoot, clicontract.ProjectRunnerVersion())
	t.Setenv(nativepath.CacheDirEnvName, t.TempDir())
	t.Setenv(dispatcherDisableSelfUpdateEnvName, "1")
	t.Chdir(projectRoot)

	deps := defaultDispatcherRunDeps()
	deps.stdoutIsTerminal = alwaysTerminalStdout()
	deps.runRealCLI = func(context.Context, string, []string, io.Writer, io.Writer) int {
		t.Fatal("interactive version request must not forward to the pinned project runner")
		return 0
	}

	var stdout bytes.Buffer
	code := runDispatcherWithDeps(context.Background(), []string{"--version"}, &stdout, io.Discard, deps)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := dispatcherVersion + "\n" +
		"This Unity project pins tloop project runner " + clicontract.ProjectRunnerVersion() + ".\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunDispatcherPrintsNoProjectContextWhenStdoutIsTerminal(t *testing.T) {
	// Verifies an interactive --version outside any Unity project explains that no project was detected.
	t.Chdir(t.TempDir())

	deps := defaultDispatcherRunDeps()
	deps.stdoutIsTerminal = alwaysTerminalStdout()

	var stdout bytes.Buffer
	code := runDispatcherWithDeps(context.Background(), []string{"--version"}, &stdout, io.Discard, deps)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := dispatcherVersion + "\n" +
		"No Unity project detected here; run inside a Unity project to see which project runner serves it.\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunDispatcherPrintsMissingPinContextWhenStdoutIsTerminal(t *testing.T) {
	// Verifies an interactive --version in a pinless V3 project points at the pin file Unity must create.
	projectRoot := createDispatcherUnityProject(t)
	t.Chdir(projectRoot)

	deps := defaultDispatcherRunDeps()
	deps.stdoutIsTerminal = alwaysTerminalStdout()

	var stdout bytes.Buffer
	code := runDispatcherWithDeps(context.Background(), []string{"--version"}, &stdout, io.Discard, deps)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := dispatcherVersion + "\n" +
		"This Unity project has no readable project runner pin yet; open it in Unity once to create " +
		dispatcherProjectPinRelativePath + ".\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunDispatcherKeepsVersionJSONMachineReadableWhenStdoutIsTerminal(t *testing.T) {
	// Verifies --version --json stays the machine payload even on a terminal, since only bare --version gains human output.
	t.Chdir(t.TempDir())

	deps := defaultDispatcherRunDeps()
	deps.stdoutIsTerminal = alwaysTerminalStdout()

	var stdout bytes.Buffer
	code := runDispatcherWithDeps(context.Background(), []string{"--version", "--json"}, &stdout, io.Discard, deps)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	payload := map[string]string{}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("parse version JSON: %v; stdout=%s", err, stdout.String())
	}
	if payload["DispatcherVersion"] != dispatcherVersion {
		t.Fatalf("DispatcherVersion = %q, want %q", payload["DispatcherVersion"], dispatcherVersion)
	}
}

func TestIsTerminalWriterRejectsNonFileWriter(t *testing.T) {
	// Verifies the default terminal probe treats an in-memory writer as non-interactive.
	if isTerminalWriter(&bytes.Buffer{}) {
		t.Fatal("a bytes.Buffer must not be reported as a terminal")
	}
}

func TestDispatcherStdoutIsTerminalTreatsMissingProbeAndFilesAsNonTerminal(t *testing.T) {
	// Verifies deps without a terminal probe, and a regular file as stdout, both count as non-terminal.
	if dispatcherStdoutIsTerminal(io.Discard, dispatcherRunDeps{}) {
		t.Fatal("deps without a probe must not count as a terminal")
	}
	file, err := os.Create(filepath.Join(t.TempDir(), "stdout.txt"))
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })
	if isTerminalWriter(file) {
		t.Fatal("a regular file must not count as a terminal")
	}
}
