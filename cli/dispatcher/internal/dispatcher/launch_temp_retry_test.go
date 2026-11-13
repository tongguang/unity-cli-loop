package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestTempRetryStopsOnWrappedPermanentFailure(t *testing.T) {
	// Verifies a permanent failure retains its cause through wrapping and stops without sleeping.
	cause := errors.New("lockfile is held")
	wrapped := fmt.Errorf("delete lockfile: %w", unretryableTempError{cause: cause})
	if !errors.Is(wrapped, cause) || wrapped.Error() != "delete lockfile: lockfile is held" {
		t.Fatalf("wrapped failure lost its cause: %v", wrapped)
	}
	deps := defaultLaunchDeps()
	deps.sleep = func(time.Duration) { t.Fatal("a permanent failure must not sleep") }
	calls := 0
	err := retryWhileTempFileIsHeld(context.Background(), deps, func() error {
		calls++
		return wrapped
	})
	if err != cause || calls != 1 {
		t.Fatalf("got error %v after %d attempts, want the original cause after one attempt", err, calls)
	}
}

func TestTempCleanupPreservesPermanentHeldLockfileFailure(t *testing.T) {
	// Verifies cleanup preserves the held-project error and leaves Temp untouched on every platform.
	projectRoot := createLaunchTestProject(t)
	createStaleUnityLockfile(t, projectRoot)
	lockfilePath := unityLockfilePath(projectRoot)
	deps := defaultLaunchDeps()
	deps.isUnityLockfileHeld = func(string) (bool, error) { return false, nil }
	deps.sleep = func(time.Duration) { t.Fatal("a held lockfile must not be retried") }
	deps.removePath = func(path string) error {
		if path != lockfilePath {
			t.Fatalf("cleanup must not remove Temp after a held lockfile: %s", path)
		}
		return unretryableTempError{cause: heldUnityLockfileError{lockfilePath: path}}
	}

	_, err := cleanStaleUnityTemp(context.Background(), projectRoot, deps)
	var held heldUnityLockfileError
	if !errors.As(err, &held) || held.lockfilePath != lockfilePath {
		t.Fatalf("expected the original held-lockfile error, got %v", err)
	}
	if _, err := os.Stat(lockfilePath); err != nil {
		t.Fatalf("cleanup removed the held lockfile: %v", err)
	}
}
