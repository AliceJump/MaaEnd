//go:build !windows

package gamelaunch

import "errors"

func run(_ []string) error {
	return errors.New("GameLaunch is only supported on Windows")
}
