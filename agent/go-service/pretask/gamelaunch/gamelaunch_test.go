package gamelaunch

import "testing"

func TestParseOptions(t *testing.T) {
	opts, err := parseOptions([]string{`{"GameLaunch":{"GamePath":"D:\\Game Dir\\Endfield.exe","LauncherPath":"D:\\Launcher\\Launcher.exe","Arguments":"-launcher_lang=en-us -launcher_sub_channel=7"}}`})
	if err != nil || opts.GamePath != `D:\Game Dir\Endfield.exe` || opts.LauncherPath != `D:\Launcher\Launcher.exe` || opts.Arguments != "-launcher_lang=en-us -launcher_sub_channel=7" {
		t.Fatalf("unexpected PI input options: %+v, %v", opts, err)
	}
	for _, raw := range []string{`{}`, `{"GameLaunch":{"GamePath":""}}`} {
		opts, err := parseOptions([]string{raw})
		if err != nil || opts.Arguments != defaultArguments {
			t.Fatalf("missing arguments should use CN defaults: %+v, %v", opts, err)
		}
	}
	opts, err = parseOptions([]string{`{"GameLaunch":{"Arguments":""}}`})
	if err != nil || opts.Arguments != "" {
		t.Fatalf("explicit empty arguments must be preserved: %+v, %v", opts, err)
	}
	for _, raw := range []string{`{`, `{"GameLaunch":"wrong type"}`} {
		if _, err := parseOptions([]string{raw}); err == nil {
			t.Fatalf("invalid options accepted: %s", raw)
		}
	}
}
