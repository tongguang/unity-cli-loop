package install

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCommandForWindowsConfiguresUserPath(t *testing.T) {
	// Verifies Windows install delegates User PATH setup and shadowing reports to the native setup command.
	command, err := CommandForOS("windows", Options{
		InstallDir: `C:\Users\ExampleUser\AppData\Local\Programs\tloop\bin`,
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	if command.Name != "powershell" {
		t.Fatalf("command name mismatch: %s", command.Name)
	}
	joinedArgs := strings.Join(command.Args, " ")
	if !strings.Contains(joinedArgs, "-EncodedCommand") {
		t.Fatalf("encoded command flag missing: %s", joinedArgs)
	}
	setupScript := windowsInstallScript(command.InstallDir, command.TargetPath)
	for _, expected := range []string{
		"[Environment]::SetEnvironmentVariable('Path', $NewUserPath, 'User')",
		"Report-PathShadowing",
		"function Get-FirstUloopCommandFromPath",
		"$NormalizedPathEntry = & $NormalizePath $PathEntry",
		"if ($NormalizedPathEntry -match '^[A-Za-z]:$') {\n            $NormalizedPathEntry = $NormalizedPathEntry + '\\'\n        }",
		"$CandidatePath = Join-Path $NormalizedPathEntry $ShimName",
		"$MachinePath = [Environment]::GetEnvironmentVariable('Path', 'Machine')",
		"$ResolvedPath = Get-FirstUloopCommandFromPath -PathValue ([string]::Join(';', @($MachinePath, $UserPath)))",
	} {
		if !strings.Contains(setupScript, expected) {
			t.Fatalf("expected %s in setup script: %s", expected, setupScript)
		}
	}
	if !command.UpdatesPath {
		t.Fatal("windows install should update User PATH")
	}
	if command.TargetPath != `C:\Users\ExampleUser\AppData\Local\Programs\tloop\bin\tloop.exe` {
		t.Fatalf("target path mismatch: %s", command.TargetPath)
	}
	if strings.Contains(setupScript, "Get-Command tloop -ErrorAction SilentlyContinue") {
		t.Fatal("Windows path shadowing should inspect persisted PATH instead of the current process PATH")
	}
	if strings.Contains(setupScript, "npm") {
		t.Fatal("install setup must not run npm: no tloop npm package exists, so an uninstall could only hit a package this project does not own")
	}
}

func TestCommandForMacConfiguresShellPath(t *testing.T) {
	// Verifies macOS install delegates shell PATH setup and shadowing reports to the native setup command.
	command, err := CommandForOS("darwin", Options{
		InstallDir: "/Users/ExampleUser/.local/bin",
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	if command.Name != "sh" {
		t.Fatalf("command name mismatch: %s", command.Name)
	}
	joinedArgs := strings.Join(command.Args, " ")
	if !strings.Contains(joinedArgs, "-c") {
		t.Fatalf("shell command flag missing: %s", joinedArgs)
	}
	setupScript := posixInstallScript(command.InstallDir, command.TargetPath)
	for _, expected := range []string{
		"# >>> tloop PATH >>>",
		"# <<< tloop PATH <<<",
		"fish_add_path --move",
		`tmp_path=$(mktemp "$profile_dir/.tloop_path.XXXXXX" 2>/dev/null || true)`,
		"report_path_shadowing",
	} {
		if !strings.Contains(setupScript, expected) {
			t.Fatalf("expected %s in setup script: %s", expected, setupScript)
		}
	}
	if !command.UpdatesPath {
		t.Fatal("macOS install should update shell PATH")
	}
	if command.TargetPath != "/Users/ExampleUser/.local/bin/tloop" {
		t.Fatalf("target path mismatch: %s", command.TargetPath)
	}
	if strings.Contains(setupScript, "npm") {
		t.Fatal("install setup must not run npm: no tloop npm package exists, so an uninstall could only hit a package this project does not own")
	}
}

func TestCommandForLinuxConfiguresShellPath(t *testing.T) {
	// Verifies Linux install uses the same POSIX setup command as macOS, including PATH setup.
	command, err := CommandForOS("linux", Options{
		InstallDir: "/home/tester/.local/bin",
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	if command.Name != "sh" {
		t.Fatalf("command name mismatch: %s", command.Name)
	}
	if command.TargetPath != "/home/tester/.local/bin/tloop" {
		t.Fatalf("target path mismatch: %s", command.TargetPath)
	}
	if !command.UpdatesPath {
		t.Fatal("Linux install should update shell PATH")
	}
}

func TestWindowsInstallScriptReplacesTemplateValues(t *testing.T) {
	// Verifies Windows setup templates cannot ship with unresolved placeholders.
	installDir := `C:\Temp\tloop's bin`
	targetPath := `C:\Temp\tloop's bin\tloop.exe`
	setupScript := windowsInstallScript(installDir, targetPath)

	if strings.Contains(setupScript, "{{") {
		t.Fatalf("setup script contains unresolved template placeholder: %s", setupScript)
	}
	if !strings.Contains(setupScript, `$InstallDir = 'C:\Temp\tloop''s bin'`) {
		t.Fatalf("setup script does not quote install dir correctly: %s", setupScript)
	}
	if !strings.Contains(setupScript, `$ExpectedUloopPath = 'C:\Temp\tloop''s bin\tloop.exe'`) {
		t.Fatalf("setup script does not quote target path correctly: %s", setupScript)
	}
}

func TestPosixInstallScriptReplacesTemplateValues(t *testing.T) {
	// Verifies POSIX setup templates cannot ship with unresolved placeholders.
	installDir := "/tmp/tloop's bin"
	targetPath := "/tmp/tloop's bin/tloop"
	setupScript := posixInstallScript(installDir, targetPath)

	if strings.Contains(setupScript, "{{") {
		t.Fatalf("setup script contains unresolved template placeholder: %s", setupScript)
	}
	if !strings.Contains(setupScript, `InstallDir='/tmp/tloop'"'"'s bin'`) {
		t.Fatalf("setup script does not quote install dir correctly: %s", setupScript)
	}
	if !strings.Contains(setupScript, `ExpectedUloopPath='/tmp/tloop'"'"'s bin/tloop'`) {
		t.Fatalf("setup script does not quote target path correctly: %s", setupScript)
	}
}

func TestCommandForMacPreservesRootInstallDir(t *testing.T) {
	// Verifies macOS install keeps the root directory when removing trailing separators.
	command, err := CommandForOS("darwin", Options{
		InstallDir: "/",
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	if command.InstallDir != "/" {
		t.Fatalf("install dir mismatch: %s", command.InstallDir)
	}
	if command.TargetPath != "/tloop" {
		t.Fatalf("target path mismatch: %s", command.TargetPath)
	}
}

func shellProfileQuoteForTest(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func TestPosixInstallScriptWritesZshPathBlock(t *testing.T) {
	// Verifies macOS shell setup writes a replaceable profile block for future terminals.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell setup is not available on Windows")
	}

	home := t.TempDir()
	installDir := filepath.Join(home, ".local", "bin")
	command, err := CommandForOS("darwin", Options{
		InstallDir: installDir,
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	process := exec.Command(command.Name, command.Args...)
	process.Env = []string{
		"HOME=" + home,
		"ZDOTDIR=" + home,
		"SHELL=/bin/zsh",
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
	output, err := process.CombinedOutput()
	if err != nil {
		t.Fatalf("POSIX setup failed: %v\n%s", err, output)
	}

	profileContent, err := os.ReadFile(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatalf("failed to read zsh profile: %v", err)
	}
	content := string(profileContent)
	for _, expected := range []string{
		"# >>> tloop PATH >>>",
		"export PATH=" + shellProfileQuoteForTest(installDir) + ":$PATH",
		"# <<< tloop PATH <<<",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("profile content missing %q:\n%s", expected, content)
		}
	}
}

func TestPosixInstallScriptWritesBashrcOnLinux(t *testing.T) {
	// Verifies Linux bash setup writes ~/.bashrc, because Linux terminals start non-login shells that skip ~/.bash_profile.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell setup is not available on Windows")
	}

	home := t.TempDir()
	installDir := filepath.Join(home, ".local", "bin")
	if err := os.WriteFile(filepath.Join(home, ".bash_profile"), []byte("existing\n"), 0o600); err != nil {
		t.Fatalf("failed to write bash profile: %v", err)
	}
	fakeBinDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(fakeBinDir, "uname"), []byte("#!/bin/sh\necho Linux\n"), 0o755); err != nil {
		t.Fatalf("failed to write fake uname: %v", err)
	}
	command, err := CommandForOS("linux", Options{
		InstallDir: installDir,
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	process := exec.Command(command.Name, command.Args...)
	process.Env = []string{
		"HOME=" + home,
		"SHELL=/bin/bash",
		"PATH=" + fakeBinDir + ":/usr/bin:/bin:/usr/sbin:/sbin",
	}
	output, err := process.CombinedOutput()
	if err != nil {
		t.Fatalf("POSIX setup failed: %v\n%s", err, output)
	}

	bashrcContent, err := os.ReadFile(filepath.Join(home, ".bashrc"))
	if err != nil {
		t.Fatalf("failed to read bashrc: %v\n%s", err, output)
	}
	if !strings.Contains(string(bashrcContent), "export PATH="+shellProfileQuoteForTest(installDir)+":$PATH") {
		t.Fatalf("bashrc missing tloop PATH block:\n%s", bashrcContent)
	}
}

func TestPosixInstallScriptCreatesNestedFishProfile(t *testing.T) {
	// Verifies macOS shell setup creates nested shell profile directories when needed.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell setup is not available on Windows")
	}

	home := t.TempDir()
	installDir := filepath.Join(home, ".local", "bin")
	command, err := CommandForOS("darwin", Options{
		InstallDir: installDir,
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	process := exec.Command(command.Name, command.Args...)
	process.Env = []string{
		"HOME=" + home,
		"SHELL=/opt/homebrew/bin/fish",
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
	output, err := process.CombinedOutput()
	if err != nil {
		t.Fatalf("POSIX setup failed: %v\n%s", err, output)
	}

	profilePath := filepath.Join(home, ".config", "fish", "config.fish")
	profileContent, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("failed to read fish profile: %v\n%s", err, output)
	}
	content := string(profileContent)
	for _, expected := range []string{
		"# >>> tloop PATH >>>",
		"fish_add_path --move " + shellProfileQuoteForTest(installDir),
		"# <<< tloop PATH <<<",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("profile content missing %q:\n%s", expected, content)
		}
	}
}

func TestPosixInstallScriptWritesThroughSymlinkedProfile(t *testing.T) {
	// Verifies macOS shell setup preserves managed shell profile symlinks.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell setup is not available on Windows")
	}

	home := t.TempDir()
	dotfilesDir := filepath.Join(home, "dotfiles")
	if err := os.MkdirAll(dotfilesDir, 0o755); err != nil {
		t.Fatalf("failed to create dotfiles directory: %v", err)
	}
	targetProfile := filepath.Join(dotfilesDir, "zshrc")
	if err := os.WriteFile(targetProfile, []byte("export EXISTING_PROFILE=1\n"), 0o644); err != nil {
		t.Fatalf("failed to write target profile: %v", err)
	}
	profilePath := filepath.Join(home, ".zshrc")
	if err := os.Symlink(filepath.Join("dotfiles", "zshrc"), profilePath); err != nil {
		t.Fatalf("failed to create profile symlink: %v", err)
	}

	installDir := filepath.Join(home, ".local", "bin")
	command, err := CommandForOS("darwin", Options{
		InstallDir: installDir,
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	process := exec.Command(command.Name, command.Args...)
	process.Env = []string{
		"HOME=" + home,
		"ZDOTDIR=" + home,
		"SHELL=/bin/zsh",
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
	output, err := process.CombinedOutput()
	if err != nil {
		t.Fatalf("POSIX setup failed: %v\n%s", err, output)
	}

	profileInfo, err := os.Lstat(profilePath)
	if err != nil {
		t.Fatalf("failed to stat profile symlink: %v", err)
	}
	if profileInfo.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("profile symlink was replaced: %s", profileInfo.Mode())
	}
	profileContent, err := os.ReadFile(targetProfile)
	if err != nil {
		t.Fatalf("failed to read target profile: %v", err)
	}
	content := string(profileContent)
	for _, expected := range []string{
		"export EXISTING_PROFILE=1",
		"export PATH=" + shellProfileQuoteForTest(installDir) + ":$PATH",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("profile target missing %q:\n%s", expected, content)
		}
	}
}

func TestPosixInstallScriptPreservesProfileWhenFilteringFails(t *testing.T) {
	// Verifies macOS shell setup does not replace an existing profile when its current content cannot be read.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell setup is not available on Windows")
	}

	home := t.TempDir()
	mockBin := filepath.Join(home, "mock-bin")
	if err := os.MkdirAll(mockBin, 0o755); err != nil {
		t.Fatalf("failed to create mock bin: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mockBin, "awk"), []byte("#!/bin/sh\nexit 7\n"), 0o755); err != nil {
		t.Fatalf("failed to write mock awk: %v", err)
	}
	profilePath := filepath.Join(home, ".zshrc")
	existingProfile := "export EXISTING_PROFILE=1\n"
	if err := os.WriteFile(profilePath, []byte(existingProfile), 0o644); err != nil {
		t.Fatalf("failed to write existing profile: %v", err)
	}

	command, err := CommandForOS("darwin", Options{
		InstallDir: filepath.Join(home, ".local", "bin"),
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	process := exec.Command(command.Name, command.Args...)
	process.Env = []string{
		"HOME=" + home,
		"ZDOTDIR=" + home,
		"SHELL=/bin/zsh",
		"PATH=" + mockBin + ":/usr/bin:/bin:/usr/sbin:/sbin",
	}
	output, err := process.CombinedOutput()
	if err == nil {
		t.Fatalf("expected POSIX setup failure:\n%s", output)
	}
	if !strings.Contains(string(output), "Could not read existing shell profile") {
		t.Fatalf("setup failure did not include profile read error:\n%s", output)
	}
	profileContent, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("failed to read profile after setup failure: %v", err)
	}
	if string(profileContent) != existingProfile {
		t.Fatalf("profile should remain unchanged:\n%s", profileContent)
	}
}

func TestPosixInstallScriptEscapesInstallDirInProfiles(t *testing.T) {
	// Verifies macOS shell setup writes literal install paths into persisted shell profile commands.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell setup is not available on Windows")
	}

	home := t.TempDir()
	installDir := filepath.Join(home, "bin with $HOME \"quote\" `date` 'single' (echo hi)")
	for _, scenario := range []struct {
		name        string
		env         []string
		profilePath string
		expected    string
	}{
		{
			name: "zsh",
			env: []string{
				"HOME=" + home,
				"ZDOTDIR=" + home,
				"SHELL=/bin/zsh",
				"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
			},
			profilePath: filepath.Join(home, ".zshrc"),
			expected:    "export PATH=" + shellProfileQuoteForTest(installDir) + ":$PATH",
		},
		{
			name: "fish",
			env: []string{
				"HOME=" + home,
				"SHELL=/opt/homebrew/bin/fish",
				"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
			},
			profilePath: filepath.Join(home, ".config", "fish", "config.fish"),
			expected:    "fish_add_path --move " + shellProfileQuoteForTest(installDir),
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			command, err := CommandForOS("darwin", Options{
				InstallDir: installDir,
			})
			if err != nil {
				t.Fatalf("CommandForOS failed: %v", err)
			}

			process := exec.Command(command.Name, command.Args...)
			process.Env = scenario.env
			output, err := process.CombinedOutput()
			if err != nil {
				t.Fatalf("POSIX setup failed: %v\n%s", err, output)
			}
			profileContent, err := os.ReadFile(scenario.profilePath)
			if err != nil {
				t.Fatalf("failed to read shell profile: %v\n%s", err, output)
			}
			if !strings.Contains(string(profileContent), scenario.expected) {
				t.Fatalf("profile content missing escaped install dir %q:\n%s", scenario.expected, profileContent)
			}
		})
	}
}

func TestPosixInstallScriptFailsWhenSymlinkProfileTargetCannotBeWritten(t *testing.T) {
	// Verifies macOS shell setup reports profile write failures to the installer.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell setup is not available on Windows")
	}

	home := t.TempDir()
	profilePath := filepath.Join(home, ".zshrc")
	if err := os.Symlink(filepath.Join(string(os.PathSeparator), "dev", "null", "tloop"), profilePath); err != nil {
		t.Fatalf("failed to create profile symlink: %v", err)
	}

	command, err := CommandForOS("darwin", Options{
		InstallDir: filepath.Join(home, ".local", "bin"),
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	process := exec.Command(command.Name, command.Args...)
	process.Env = []string{
		"HOME=" + home,
		"ZDOTDIR=" + home,
		"SHELL=/bin/zsh",
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
	output, err := process.CombinedOutput()
	if err == nil {
		t.Fatalf("expected POSIX setup failure:\n%s", output)
	}
	if !strings.Contains(string(output), "Could not create shell profile directory") {
		t.Fatalf("setup failure did not include profile directory error:\n%s", output)
	}
}

func TestPosixInstallScriptSucceedsWhenShellProfileIsUnknown(t *testing.T) {
	// Verifies macOS shell setup treats manual PATH guidance as non-fatal.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell setup is not available on Windows")
	}

	home := t.TempDir()
	installDir := filepath.Join(home, ".local", "bin")
	command, err := CommandForOS("darwin", Options{
		InstallDir: installDir,
	})
	if err != nil {
		t.Fatalf("CommandForOS failed: %v", err)
	}

	process := exec.Command(command.Name, command.Args...)
	process.Env = []string{
		"HOME=" + home,
		"SHELL=/bin/tcsh",
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
	output, err := process.CombinedOutput()
	if err != nil {
		t.Fatalf("expected POSIX setup success:\n%s", output)
	}
	outputText := string(output)
	if !strings.Contains(outputText, "Add this directory to PATH in your shell profile:") {
		t.Fatalf("setup output should include manual PATH guidance:\n%s", outputText)
	}
	if !strings.Contains(outputText, installDir) {
		t.Fatalf("manual PATH guidance should include install dir:\n%s", outputText)
	}
}

func TestCommandForOSRejectsUnsupportedOS(t *testing.T) {
	// Verifies unsupported platforms fail before building any setup command.
	_, err := CommandForOS("freebsd", Options{
		InstallDir: "/Users/ExampleUser/.local/bin",
	})
	if err == nil {
		t.Fatal("expected unsupported OS error")
	}
	if !strings.Contains(err.Error(), "macOS, Linux, and Windows") {
		t.Fatalf("unexpected unsupported OS error: %v", err)
	}
}

func TestWindowsInstallScriptParsesOnWindows(t *testing.T) {
	// Verifies the embedded setup script remains valid PowerShell on Windows.
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell parser check is Windows-only")
	}

	setupScript := windowsInstallScript(
		`C:\Users\ExampleUser\AppData\Local\Programs\tloop\bin`,
		`C:\Users\ExampleUser\AppData\Local\Programs\tloop\bin\tloop.exe`)
	scriptPath := t.TempDir() + `\install-setup.ps1`
	if err := os.WriteFile(scriptPath, []byte(setupScript), 0o600); err != nil {
		t.Fatalf("failed to write setup script: %v", err)
	}

	// PowerShell's cold start on Windows CI runners exceeds 10 seconds often
	// enough that a tighter bound fails this parse check without a parse error.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// The script path must be embedded in the command text: powershell -Command
	// concatenates trailing arguments into the command instead of exposing them
	// via $args, so a trailing path argument would be invoked as a command and
	// execute the installer setup for real.
	parseCommand := `$parseErrors = $null; $null = [System.Management.Automation.PSParser]::Tokenize((Get-Content -Raw -LiteralPath ` +
		powerShellSingleQuote(scriptPath) +
		`), [ref]$parseErrors); if ($parseErrors) { $parseErrors | Out-String; exit 1 }`
	command := exec.CommandContext(
		ctx,
		"powershell",
		"-NoProfile",
		"-Command",
		parseCommand)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("embedded setup script does not parse: %v\n%s", err, output)
	}
}
