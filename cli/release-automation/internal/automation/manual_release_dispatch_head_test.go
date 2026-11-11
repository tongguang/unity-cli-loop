package automation

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// Verifies transient old API responses cannot dispatch a workflow before two consecutive matching reads.
func TestManualReleaseDispatchWaitsForStableAPIHead(t *testing.T) {
	fixture := newManualReleaseFixture(t)
	deps := fixture.deps()
	responses := []string{"old", "source", "old", "source", "source"}
	reads, waits := 0, 0
	deps.run = func(ctx context.Context, name string, args ...string) (string, error) {
		if name == "gh" && args[0] == "api" && strings.HasSuffix(args[1], "/main") {
			response := responses[reads]
			reads++
			return response, nil
		}
		return fixture.run(ctx, name, args...)
	}
	deps.wait = func(context.Context, time.Duration) error { waits++; return nil }
	if err := waitForManualReleaseDispatchHead(context.Background(), fixture.options(), deps, "source"); err != nil {
		t.Fatal(err)
	}
	if reads != 5 || waits != 4 {
		t.Fatalf("dispatch became ready too soon: %d reads, %d waits", reads, waits)
	}
}

// Verifies stale API timeouts, API failures, cancellation and real branch movement stop dispatch.
func TestManualReleaseDispatchHeadFailuresStopWorkflow(t *testing.T) {
	for _, failure := range []string{"stale", "api", "cancel", "moved"} {
		t.Run(failure, func(t *testing.T) {
			fixture := newManualReleaseFixture(t)
			deps := fixture.deps()
			deps.run = func(ctx context.Context, name string, args ...string) (string, error) {
				if name == "gh" && args[0] == "api" && strings.HasSuffix(args[1], "/main") {
					if failure == "api" {
						return "", errors.New("API unavailable")
					}
					return "old", nil
				}
				return fixture.run(ctx, name, args...)
			}
			waits := 0
			deps.wait = func(context.Context, time.Duration) error {
				waits++
				if failure == "cancel" {
					return context.Canceled
				}
				if failure == "moved" {
					fixture.remote = "new-head"
				}
				return nil
			}
			err := dispatchManualReleaseWorkflow(context.Background(), io.Discard, fixture.options(), deps, "build-and-test.yml", "source", "native-ci")
			if err == nil || (failure == "stale" && waits != 12) {
				t.Fatalf("failure %s returned %v after %d waits", failure, err, waits)
			}
			if strings.Contains(strings.Join(fixture.calls, "\n"), "workflow run") {
				t.Fatal("workflow started before the head was ready")
			}
		})
	}
}
