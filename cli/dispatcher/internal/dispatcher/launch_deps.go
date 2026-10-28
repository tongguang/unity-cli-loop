package dispatcher

import (
	"context"
	"os"
	"time"

	"github.com/hatayama/unity-cli-loop/common/clicore"
	"github.com/hatayama/unity-cli-loop/common/unityprocess"
)

type launchDeps struct {
	now                        func() time.Time
	findRunningUnityProcess    func(context.Context, string) (*unityprocess.UnityProcess, error)
	focusUnityProcess          func(context.Context, int) error
	killUnityProcess           func(int) error
	requestUnityQuit           func(context.Context, string) bool
	resolveUnityExecutablePath func(string) (string, error)
	waitForUnityProcessExit    func(context.Context, string, int, time.Duration, time.Duration) error
	waitForUnityStartupMarker  func(context.Context, string, time.Duration, time.Duration) error
	waitForToolReadiness       func(context.Context, string, time.Duration) error
	probeProjectIpcFallback    func(context.Context, string) error
	removePath                 func(string) error
	statPath                   func(string) (os.FileInfo, error)
	sleep                      func(time.Duration)
}

func defaultLaunchDeps() launchDeps {
	return launchDeps{
		now:                        time.Now,
		findRunningUnityProcess:    unityprocess.FindRunningUnityProcess,
		focusUnityProcess:          unityprocess.FocusUnityProcess,
		killUnityProcess:           killUnityProcess,
		requestUnityQuit:           requestUnityQuit,
		resolveUnityExecutablePath: resolveUnityExecutablePath,
		waitForUnityProcessExit:    waitForUnityProcessExit,
		waitForUnityStartupMarker:  waitForUnityStartupMarkerOrTimeout,
		waitForToolReadiness:       clicore.WaitForToolReadinessWithTimeout,
		probeProjectIpcFallback:    clicore.ProbeToolReadinessSequence,
		removePath:                 os.RemoveAll,
		statPath:                   os.Stat,
		sleep:                      time.Sleep,
	}
}
