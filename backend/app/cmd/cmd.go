package cmd

import (
	"os"

	"github.com/pkg/errors"
)

// mkdir -p for all dirs
func makeDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil { // If path is already a directory, MkdirAll does nothing
			return errors.Wrapf(err, "can't make directory %s", dir)
		}
	}
	return nil
}
