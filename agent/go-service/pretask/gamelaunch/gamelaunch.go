package gamelaunch

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

// Options 是 GameSetting 预任务传入的国服官服游戏和启动器路径。
type Options struct {
	GamePath     string
	LauncherPath string
}

// Run 在游戏设置完成后启动国服官服；已有游戏进程时不重复启动。
func Run(opts Options) bool {
	if err := run(opts); err != nil {
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

func executablePath(path, name string) (string, error) {
	path = strings.TrimSpace(path)
	if !filepath.IsAbs(path) || !strings.EqualFold(filepath.Base(path), name) {
		return "", fmt.Errorf("configure an absolute path to %s", name)
	}
	return filepath.Clean(path), nil
}
