// release runs the explicit, ordered Tuanjie release workflow.
package main

import (
	"context"
	"os"
	"time"

	"github.com/hatayama/unity-cli-loop/tools/release-automation/internal/automation"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()
	os.Exit(automation.RunRelease(ctx, os.Stdout, os.Stderr, os.Args[1:]))
}
