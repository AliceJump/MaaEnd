package gamelaunch

import "testing"

func TestParseOptions(t *testing.T) {
	opts, err := parseOptions([]string{`{"GameLaunch":{"GamePath":"D:\\Game Dir\\Endfield.exe","LauncherPath":"D:\\Launcher\\Launcher.exe"}}`})
	if err != nil || opts.GamePath != `D:\Game Dir\Endfield.exe` || opts.LauncherPath != `D:\Launcher\Launcher.exe` {
		t.Fatalf("unexpected PI input options: %+v, %v", opts, err)
	}
	for _, raw := range []string{`{`, `{"GameLaunch":"wrong type"}`} {
		if _, err := parseOptions([]string{raw}); err == nil {
			t.Fatalf("invalid options accepted: %s", raw)
		}
	}
}
