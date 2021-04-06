package util

import (
	"os/user"
	"path/filepath"
	"strings"
)

func ResolvePath(path string) string {
	usr, err := user.Current()
	if err != nil {
		return path
	}

	switch {
	case path == "~":
		return usr.HomeDir
	case strings.HasPrefix(path, "~/"):
		return filepath.Join(usr.HomeDir, path[2:])
	default:
		return path
	}
}
