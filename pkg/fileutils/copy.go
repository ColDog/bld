package fileutils

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Copy will copy a directory.
func Copy(src, dest string) error {
	if err := os.MkdirAll(dest, Directory); err != nil {
		return err
	}
	src = filepath.Clean(src)
	dest = filepath.Clean(dest)
	cmd := exec.Command("cp", "-r", src+"/.", dest+"/")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
