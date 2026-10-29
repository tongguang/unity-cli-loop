package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type manualReleaseWorkflowRun struct {
	DatabaseID   int64  `json:"databaseId"`
	HeadSHA      string `json:"headSha"`
	DisplayTitle string `json:"displayTitle"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	URL          string `json:"url"`
}

func dispatchManualReleaseWorkflow(ctx context.Context, out io.Writer, options manualReleaseOptions, deps manualReleaseDeps, workflow, sha, phase string) error {
	if err := requireManualReleaseHead(ctx, options, sha, deps.run); err != nil {
		return err
	}
	id := "tloop-release-" + options.RunID + "-" + phase
	args := []string{"workflow", "run", workflow, "--repo", options.Repository, "--ref", "main", "-f", "expected-sha=" + sha, "-f", "release-id=" + id}
	if workflow != "build-and-test.yml" {
		args = append(args, "-f", "dry-run=false")
	}
	if _, err := deps.run(ctx, "gh", args...); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, "Waiting for", workflow, id, "at", sha)
	for range 360 {
		done, err := checkManualReleaseWorkflow(ctx, options, deps.run, workflow, sha, id)
		if err != nil || done {
			return err
		}
		if err := deps.wait(ctx, 10*time.Second); err != nil {
			return err
		}
	}
	return fmt.Errorf("timed out waiting for %s (%s)", workflow, id)
}

func checkManualReleaseWorkflow(ctx context.Context, options manualReleaseOptions, run manualReleaseRun, workflow, sha, id string) (bool, error) {
	output, err := run(ctx, "gh", "run", "list", "--repo", options.Repository, "--workflow", workflow, "--branch", "main", "--event", "workflow_dispatch", "--json", "databaseId,headSha,displayTitle,status,conclusion,url", "--limit", "50")
	if err != nil {
		return false, err
	}
	var runs []manualReleaseWorkflowRun
	if err := json.Unmarshal([]byte(output), &runs); err != nil {
		return false, err
	}
	for _, candidate := range runs {
		if candidate.DisplayTitle != id {
			continue
		}
		return manualReleaseWorkflowResult(candidate, sha)
	}
	return false, nil
}

func manualReleaseWorkflowResult(run manualReleaseWorkflowRun, sha string) (bool, error) {
	if run.HeadSHA != sha {
		return false, fmt.Errorf("workflow %s ran at %s instead of %s", run.URL, run.HeadSHA, sha)
	}
	if run.Status != "completed" {
		return false, nil
	}
	if run.Conclusion != "success" {
		return false, fmt.Errorf("workflow failed (%s): %s", run.Conclusion, run.URL)
	}
	return true, nil
}
