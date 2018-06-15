package store

import (
	"os"
	"os/exec"
)

func tarTo(src, dest string) error {
	cmd := exec.Command("tar", "-zcf", dest, "-C", src, ".")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func untarTo(src, dest string) error {
	cmd := exec.Command("tar", "-zxf", src, "-C", dest)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
