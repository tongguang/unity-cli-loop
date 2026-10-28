package nativepath

import "testing"

// Verifies original Unity environment overrides cannot redirect Tuanjie's install or cache.
func TestTloopIgnoresUloopEnvironment(t *testing.T) {
	environment := Environment{Getenv: func(name string) string {
		switch name {
		case "ULOOP_INSTALL_DIR", "ULOOP_CACHE_DIR", "ULOOP_PROJECT_RUNNER_PATH":
			return `C:\Unity\uloop`
		case "LOCALAPPDATA":
			return `C:\Users\开发者\AppData\Local`
		default:
			return ""
		}
	}}
	installDir, err := ResolveInstallDir("windows", "", environment)
	if err != nil || installDir != `C:\Users\开发者\AppData\Local\Programs\tloop\bin` {
		t.Fatalf("unexpected Tuanjie install directory: %q, %v", installDir, err)
	}
	cacheDir, err := CacheRoot("windows", environment)
	if err != nil || cacheDir != `C:\Users\开发者\AppData\Local\tloop` {
		t.Fatalf("unexpected Tuanjie cache directory: %q, %v", cacheDir, err)
	}
}
