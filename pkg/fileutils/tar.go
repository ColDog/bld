package fileutils

import (
	"os"
	"os/exec"
)

// Tar will shell out to GNU tar to tar up a directory.
func Tar(src, dest string) error {
	cmd := exec.Command("tar", "-zcf", dest, "-C", src, ".")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Untar will shell out to GNU tar to untar a directory.
func Untar(src, dest string) error {
	cmd := exec.Command("tar", "-zxf", src, "-C", dest)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
