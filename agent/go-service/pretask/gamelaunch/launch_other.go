//go:build !windows

package gamelaunch

import "errors"

func run(_ Options) error {
	return errors.New("game startup is only supported on Windows")
}
