package dispatcher

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hatayama/unity-cli-loop/dispatcher/internal/nativepath"
	"github.com/hatayama/unity-cli-loop/dispatcher/internal/update"
)

func TestVersionHoldRequiresAResolvableConfigDirectory(t *testing.T) {
	// Verifies an unavailable user config blocks a version hold before installation, but does not block a plain update.
	for _, target := range []string{"999.0.0", dispatcherVersion, ""} {
		t.Run(target, func(t *testing.T) {
			fixture := newVersionHoldUpdateFixture(t)
			for _, name := range []string{nativepath.ConfigDirEnvName, "HOME", "USERPROFILE", "APPDATA", "XDG_CONFIG_HOME"} {
				t.Setenv(name, "")
			}
			var stderr bytes.Buffer
			code := runDispatcherUpdateRequestForOS(context.Background(), updateOptions{targetVersion: target}, "linux", io.Discard, &stderr)
			if target == "" {
				if code != 0 || fixture.installerCalls != 1 {
					t.Fatalf("plain update failed: code=%d installs=%d stderr=%s", code, fixture.installerCalls, stderr.String())
				}
				return
			}
			if code != 1 || fixture.installerCalls != 0 || !strings.Contains(stderr.String(), "Could not record the version hold") {
				t.Fatalf("hold must fail before installing: code=%d installs=%d stderr=%s", code, fixture.installerCalls, stderr.String())
			}
		})
	}
}

func TestPeriodicUpdatePropagatesLookupAndPlatformFailures(t *testing.T) {
	// Verifies release lookup and unsupported-platform errors never report an installed release or invoke the installer.
	for _, platform := range []string{"linux", "plan9"} {
		t.Run(platform, func(t *testing.T) {
			fixture := newVersionHoldUpdateFixture(t)
			if platform == "linux" {
				fixture.resolveErr = errors.New("release lookup unavailable")
			}
			installed, err := runPeriodicDispatcherUpdateForOS(context.Background(), io.Discard, platform)
			if err == nil || installed != "" || fixture.installerCalls != 0 {
				t.Fatalf("unexpected install: version=%q err=%v calls=%d", installed, err, fixture.installerCalls)
			}
		})
	}
}

func TestRequiredUpdateRejectsAnInvalidFloorRelease(t *testing.T) {
	// Verifies a valid channel response cannot allow an invalid floor lookup result to reach the installer.
	installedTags := stubRequiredUpdateRelease(t, "1.0.0")
	resolveUpdateTargetVersionFunc = func(_ context.Context, options update.Options) (update.Options, error) {
		if options.TargetVersion == "" {
			options.TargetVersion = "1.0.0"
		} else {
			options.TargetVersion = "invalid"
		}
		return options, nil
	}
	installed, err := runRequiredDispatcherUpdateForOS(context.Background(), "999.0.0", "linux")
	if err == nil || installed != "" || len(*installedTags) != 0 {
		t.Fatalf("invalid floor was installed: version=%q err=%v tags=%v", installed, err, *installedTags)
	}
}

func TestUnusableLockFileSkipsPeriodicButAllowsManualUpdate(t *testing.T) {
	// Verifies an unusable lock path skips optional updates without preventing a manual recovery update.
	fixture := newPeriodicUpdateFixture(t)
	if err := os.Mkdir(filepath.Join(fixture.cacheRoot, dispatcherUpdateLockFileName), 0o755); err != nil {
		t.Fatal(err)
	}
	fixture.run()
	if fixture.periodicCalls != 0 {
		t.Fatal("periodic update ran without an available lock")
	}
	manual := newVersionHoldUpdateFixture(t)
	code, _, stderr := runUpdateForHoldTest()
	if code != 0 || manual.installerCalls != 1 {
		t.Fatalf("manual recovery failed: code=%d installs=%d stderr=%s", code, manual.installerCalls, stderr)
	}
}

func TestUnreadableHoldPathPreventsPeriodicUpdate(t *testing.T) {
	// Verifies a directory occupying the hold path is treated as a hold instead of silently enabling updates.
	fixture := newPeriodicUpdateFixture(t)
	if err := os.Mkdir(filepath.Join(fixture.configRoot, dispatcherVersionHoldFileName), 0o755); err != nil {
		t.Fatal(err)
	}
	fixture.run()
	held, version := readDispatcherVersionHold()
	if !held || version != dispatcherVersionHoldUnknownVersion || fixture.periodicCalls != 0 {
		t.Fatalf("unreadable hold ignored: held=%t version=%q updates=%d", held, version, fixture.periodicCalls)
	}
}

func TestHoldWriteFailsWhenTemporaryFileCannotBeCreated(t *testing.T) {
	// Verifies a missing destination directory leaves no hold or temporary file after a failed write.
	configRoot := t.TempDir()
	holdPath := filepath.Join(configRoot, "missing", dispatcherVersionHoldFileName)
	if err := replaceDispatcherVersionHoldFile(holdPath, []byte(`{"version":"3.1.0"}`)); err == nil {
		t.Fatal("expected a failure when the destination directory is missing")
	}
	entries, err := os.ReadDir(configRoot)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed write left files: entries=%v err=%v", entries, err)
	}
}
