package fileutils

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Copy will copy a directory.
func Copy(src, dest string, files []string) error {
	if err := os.MkdirAll(dest, Directory); err != nil {
		return err
	}
	src = filepath.Clean(src)
	dest = filepath.Clean(dest)
	if len(files) > 0 {
		args := []string{"-r"}
		for _, f := range files {
			args = append(args, filepath.Join(src, f))
		}
		args = append(args, dest)
		cmd := exec.Command("cp", args...)
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	cmd := exec.Command("cp", "-r", src+"/.", dest+"/.")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CopyStream handles multiple read and write closers.
func CopyStream(src io.ReadCloser, dst io.WriteCloser) (err error) {
	defer func() {
		if serr := src.Close(); serr != nil {
			err = serr
		}
	}()
	defer func() {
		if ferr := dst.Close(); ferr != nil {
			err = ferr
		}
	}()

	_, err = io.Copy(dst, src)
	return err
}
