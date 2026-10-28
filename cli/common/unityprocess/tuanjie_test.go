package unityprocess

import (
	"encoding/base64"
	"strings"
	"testing"
)

// Verifies Windows enumeration requests only the Tuanjie Editor.
func TestWindowsProcessListIncludesTuanjie(t *testing.T) {
	if script := windowsUnityProcessListScript(); !strings.Contains(script, "Name = 'Tuanjie.exe'") || strings.Contains(script, "Name = 'Unity.exe'") {
		t.Fatalf("missing Editor product filter: %s", script)
	}
}

// Verifies Tuanjie process discovery preserves Chinese paths and excludes helpers and batch runs.
func TestParseWindowsTuanjieProcesses(t *testing.T) {
	projectPath := `D:\项目\团结 测试`
	executable := `"C:\Program Files\Tuanjie\Hub\Editor\2022.3.62t15\Editor\Tuanjie.exe"`
	for _, testCase := range []struct {
		name    string
		command string
		want    bool
	}{
		{"interactive", executable + ` -projectPath "` + projectPath + `" -useHub`, true},
		{"batch", executable + ` -batchmode -projectPath "` + projectPath + `"`, false},
		{"worker", executable + ` -name AssetImportWorker0 -projectPath "` + projectPath + `"`, false},
		{"hub", `"C:\Tuanjie Hub\Tuanjie Hub.exe" -projectPath "` + projectPath + `"`, false},
		{"Unity Editor", `"C:\Unity\Editor\Unity.exe" -projectPath "` + projectPath + `"`, false},
		{"different executable", `"C:\Tools\OtherTuanjie.exe" -projectPath "` + projectPath + `"`, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			output := "123|" + base64.StdEncoding.EncodeToString([]byte(testCase.command)) + "\r\n"
			got := parseWindowsUnityProcesses(output)
			if len(got) != 0 && !testCase.want || len(got) == 0 && testCase.want {
				t.Fatalf("processes = %+v, want match = %v", got, testCase.want)
			}
			if testCase.want && (got[0].Pid != 123 || got[0].projectPath != projectPath) {
				t.Fatalf("process = %+v, want project = %q", got[0], projectPath)
			}
		})
	}
}

// Verifies macOS and Linux match Tuanjie Editors while excluding Hub and import workers.
func TestMatchTuanjieProcessesOnUnix(t *testing.T) {
	projectPath := "/home/user/团结 Project"
	for _, testCase := range []struct {
		name string
		args []string
		want bool
	}{
		{"editor", []string{"/opt/Tuanjie/Editor/Tuanjie", "-projectPath", projectPath}, true},
		{"hub", []string{"/opt/TuanjieHub", "-projectPath", projectPath}, false},
		{"Unity Editor", []string{"/opt/Unity/Editor/Unity", "-projectPath", projectPath}, false},
		{"batch", []string{"Tuanjie", "-batchmode", "-projectPath", projectPath}, false},
		{"worker", []string{"Tuanjie", "-name", "AssetImportWorker0", "-projectPath", projectPath}, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, matched := matchLinuxUnityProcess(123, testCase.args)
			if matched != testCase.want || matched && got.projectPath != projectPath {
				t.Fatalf("process = %+v, matched = %v", got, matched)
			}
		})
	}
	command := `/Applications/Tuanjie/Hub/Editor/2022.3.62t15/Tuanjie.app/Contents/MacOS/Tuanjie -projectPath "` + projectPath + `"`
	got, matched := matchMacUnityProcess(123, command)
	if !matched || got.projectPath != projectPath {
		t.Fatalf("macOS process = %+v, matched = %v", got, matched)
	}
	if _, matched := matchMacUnityProcess(456, command+" -batchmode"); matched {
		t.Fatal("macOS matched a batch Editor")
	}
	if _, matched := matchMacUnityProcess(789, `/Applications/Unity.app/Contents/MacOS/Unity -projectPath /work/project`); matched {
		t.Fatal("macOS matched a Unity Editor")
	}
}
