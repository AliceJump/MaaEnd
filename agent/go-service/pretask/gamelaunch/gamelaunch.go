package gamelaunch

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

type launchOptions struct {
	GamePath     string
	LauncherPath string
}

// Run 在连接 Win32 控制器前启动国服官服；已有游戏进程时不重复启动。
func Run(args []string) bool {
	if err := run(args); err != nil {
		event := log.Error().Err(err).Str("component", "gamelaunch")
		var errno syscall.Errno
		if errors.As(err, &errno) {
			event.Uint32("windows_error", uint32(errno))
		}
		event.Msg("failed to launch game")
		return false
	}
	return true
}

func parseOptions(args []string) (launchOptions, error) {
	var opts struct{ GameLaunch launchOptions }
	if len(args) > 0 {
		if err := json.Unmarshal([]byte(args[len(args)-1]), &opts); err != nil {
			return launchOptions{}, fmt.Errorf("parse GameLaunch options: %w", err)
		}
	}
	return opts.GameLaunch, nil
}

func executablePath(path, name string) (string, error) {
	path = strings.TrimSpace(path)
	if !filepath.IsAbs(path) || !strings.EqualFold(filepath.Base(path), name) {
		return "", fmt.Errorf("configure an absolute path to %s", name)
	}
	return filepath.Clean(path), nil
}
