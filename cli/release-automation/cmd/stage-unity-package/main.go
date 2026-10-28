// stage-unity-package checks an ordinary UPM tarball and renames
// it to the release asset name OpenUPM republishes, refusing tarballs that are
// mislabeled archives or archives missing package contents.
package main

import (
	"os"

	"github.com/hatayama/unity-cli-loop/tools/release-automation/internal/automation"
)

func main() {
	os.Exit(automation.RunStageUnityPackage(os.Stdout, os.Stderr, os.Args[1:]))
}
