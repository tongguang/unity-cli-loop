package automation

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Git ref writes can reach the commits API and Actions after the push succeeds.
// Require two matching API reads before dispatch, while rejecting real branch movement immediately.
func waitForManualReleaseDispatchHead(ctx context.Context, options manualReleaseOptions, deps manualReleaseDeps, expected string) error {
	matchingReads := 0
	for range 12 {
		if err := requireManualReleaseHead(ctx, options, expected, deps.run); err != nil {
			return err
		}
		actual, err := deps.run(ctx, "gh", "api", "repos/"+options.Repository+"/commits/main", "--jq", ".sha")
		if err != nil {
			return err
		}
		if strings.TrimSpace(actual) == expected {
			matchingReads++
		} else {
			matchingReads = 0
		}
		if matchingReads == 2 {
			return nil
		}
		if err := deps.wait(ctx, 5*time.Second); err != nil {
			return err
		}
	}
	return fmt.Errorf("GitHub API did not converge to release commit %s before workflow dispatch", expected)
}
