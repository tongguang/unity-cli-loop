package update

import "testing"

// TestIsHomebrewManagedPath verifies Cellar path-segment detection across Homebrew
// prefixes and rejects substring false positives, including Windows-style paths.
func TestIsHomebrewManagedPath(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "apple silicon cellar",
			path: "/opt/homebrew/Cellar/tloop/3.0.0/bin/tloop",
			want: true,
		},
		{
			name: "intel mac cellar",
			path: "/usr/local/Cellar/tloop/3.0.0-beta.30/bin/tloop",
			want: true,
		},
		{
			name: "linuxbrew cellar",
			path: "/home/linuxbrew/.linuxbrew/Cellar/tloop/3.0.0/bin/tloop",
			want: true,
		},
		{
			name: "curl install location",
			path: "/Users/someone/.local/bin/tloop",
			want: false,
		},
		{
			name: "cellar substring segment only",
			path: "/Users/someone/MyCellarTools/tloop",
			want: false,
		},
		{
			name: "windows path with backslashes",
			path: `C:\Users\someone\AppData\Local\tloop\tloop.exe`,
			want: false,
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsHomebrewManagedPath(tc.path)
			if got != tc.want {
				t.Fatalf("IsHomebrewManagedPath(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
