package update

import "testing"

// TestIsWingetManagedPath verifies WinGet Packages and Links path detection across separators and casing.
func TestIsWingetManagedPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "user package", path: `C:\Users\<USER_NAME>\AppData\Local\Microsoft\WinGet\Packages\tongguang.tloop_Microsoft.Winget.Source_8wekyb3d8bbwe\tloop.exe`, want: true},
		{name: "user link", path: `C:\Users\<USER_NAME>\AppData\Local\Microsoft\WinGet\Links\tloop.exe`, want: true},
		{name: "machine package", path: `C:\Program Files\WinGet\Packages\tongguang.tloop_Microsoft.Winget.Source_8wekyb3d8bbwe\tloop.exe`, want: true},
		{name: "machine link", path: `C:\Program Files\WinGet\Links\tloop.exe`, want: true},
		{name: "forward slashes", path: `C:/Users/<USER_NAME>/AppData/Local/Microsoft/WinGet/Packages/tongguang.tloop_Microsoft.Winget.Source_8wekyb3d8bbwe/tloop.exe`, want: true},
		{name: "case insensitive", path: `C:\Users\<USER_NAME>\AppData\Local\Microsoft\winget\packages\tongguang.tloop\tloop.exe`, want: true},
		{name: "curl install", path: `C:\Users\<USER_NAME>\AppData\Local\Programs\tloop\bin\tloop.exe`, want: false},
		{name: "homebrew install", path: `/opt/homebrew/Cellar/tloop/3.1.0/bin/tloop`, want: false},
		{name: "winget without managed directory", path: `C:\Tools\WinGet\tloop.exe`, want: false},
		{name: "managed directory is not adjacent", path: `C:\Tools\WinGet\Other\Packages\tloop.exe`, want: false},
		{name: "reversed segments", path: `D:\Packages\WinGet\tloop.exe`, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := IsWingetManagedPath(test.path)
			if actual != test.want {
				t.Fatalf("IsWingetManagedPath(%q) = %v, want %v", test.path, actual, test.want)
			}
		})
	}
}

// TestDetectManagedInstall verifies Homebrew, winget, and unmanaged paths map to their update policies.
func TestDetectManagedInstall(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		expectedKind    ManagedInstallKind
		expectedName    string
		expectedUpgrade string
	}{
		{name: "homebrew", path: "/opt/homebrew/Cellar/tloop/3.1.0/bin/tloop", expectedKind: ManagedInstallHomebrew, expectedName: "Homebrew", expectedUpgrade: "brew upgrade tloop"},
		{name: "winget", path: `C:\Program Files\WinGet\Links\tloop.exe`, expectedKind: ManagedInstallWinget, expectedName: "winget", expectedUpgrade: "winget upgrade --id tongguang.tloop"},
		{name: "homebrew precedence", path: "/opt/homebrew/Cellar/tloop/3.1.0/WinGet/Links/tloop", expectedKind: ManagedInstallHomebrew, expectedName: "Homebrew", expectedUpgrade: "brew upgrade tloop"},
		{name: "unmanaged", path: "/home/<USER_NAME>/.local/bin/tloop", expectedKind: ManagedInstallNone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			managedInstall := DetectManagedInstall(test.path)
			if managedInstall.Kind != test.expectedKind || managedInstall.DisplayName != test.expectedName || managedInstall.UpgradeCommand != test.expectedUpgrade {
				t.Fatalf("DetectManagedInstall(%q) = %#v", test.path, managedInstall)
			}
			if managedInstall.IsManaged() != (test.expectedKind != ManagedInstallNone) {
				t.Fatalf("IsManaged() mismatch for %#v", managedInstall)
			}
		})
	}
}
