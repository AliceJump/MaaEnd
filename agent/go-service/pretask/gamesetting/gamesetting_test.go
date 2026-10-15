package gamesetting

import "testing"

func TestLaunchRemainsOptIn(t *testing.T) {
	for _, args := range [][]string{nil, {`{"GameSettingRegion":"CN","GameSettingLanguage":"Unchanged"}`}} {
		opts, err := parseGameSettingOptions(args)
		if err != nil || opts.LaunchGame != "" {
			t.Fatalf("existing settings must not enable startup: %+v, %v", opts, err)
		}
	}
}

func TestParseOptionalLaunchPath(t *testing.T) {
	opts, err := parseGameSettingOptions([]string{`{"GameSettingLaunchGame":"Yes","GameLaunch":{"GamePath":"D:\\Game Dir\\Endfield.exe"}}`})
	if err != nil || opts.Region != regionCN || opts.LaunchGame != "Yes" || opts.LaunchPaths.GamePath != `D:\Game Dir\Endfield.exe` {
		t.Fatalf("unexpected settings with startup: %+v, %v", opts, err)
	}
}

func TestLegacyLauncherPathIsIgnored(t *testing.T) {
	opts, err := parseGameSettingOptions([]string{`{"GameSettingLaunchGame":"Yes","GameLaunch":{"GamePath":"D:\\Game\\Endfield.exe","LauncherPath":"D:\\Old\\Launcher.exe"}}`})
	if err != nil || opts.LaunchPaths.GamePath != `D:\Game\Endfield.exe` {
		t.Fatalf("legacy launcher path must not break parsing: %+v, %v", opts, err)
	}
}

func TestRejectUnsupportedLaunchRegion(t *testing.T) {
	if Run([]string{`{"GameSettingRegion":"Global","GameSettingLaunchGame":"Yes"}`}) {
		t.Fatal("unverified region must be rejected before applying settings")
	}
}

func TestRunningGameWithLaunchEnabled(t *testing.T) {
	if !isGameRunning() {
		t.Skip("no running game; do not start one or change settings in tests")
	}
	if !Run([]string{`{"GameSettingLaunchGame":"Yes"}`}) {
		t.Fatal("running game must skip both settings and startup")
	}
	if Run([]string{`{"GameSettingLaunchGame":"No"}`}) {
		t.Fatal("settings-only behavior must continue rejecting a running game")
	}
}
