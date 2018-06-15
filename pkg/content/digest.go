package content

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func reader(root string, h hash.Hash) filepath.WalkFunc {
	return func(file string, info os.FileInfo, pathErr error) error {
		local, err := filepath.Rel(root, file)
		if err != nil {
			return err
		}
		_, err = h.Write([]byte(local))
		if err != nil {
			return err
		}
		if pathErr != nil {
			return pathErr
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(h, f)
		if err != nil {
			return err
		}
		return nil
	}
}

func DigestDir(root string) (string, error) {
	h := sha256.New()
	readFunc := reader(root, h)
	err := filepath.Walk(root, readFunc)
	if err != nil {
		return "", fmt.Errorf("failed digest for dir (%s): %v", root, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func DigestFiles(root string, files []string) (string, error) {
	h := sha256.New()
	readFunc := reader(root, h)
	var err error
	for _, file := range files {
		full := filepath.Join(root, file)
		info, pathErr := os.Stat(full)
		err = readFunc(full, info, pathErr)
		if err != nil {
			break
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed digest for files (%v): %v", files, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func DigestStrings(strs ...string) string {
	h := sha256.New()
	sort.Strings(strs)
	for _, str := range strs {
		h.Write([]byte(str))
	}
	return hex.EncodeToString(h.Sum(nil))
}
