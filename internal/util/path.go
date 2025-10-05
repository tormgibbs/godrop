package util

import (
	"os/user"
	"path/filepath"
)


func ExpandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = filepath.Join(usr.HomeDir, path[1:])
	}
	return path, nil
}
