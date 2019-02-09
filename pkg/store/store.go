package store

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coldog/bld/pkg/fileutils"
)

// NewLocalStore instantiates a store which will store data at dir.
func NewLocalStore(dir string) Store { return &local{dir: dir} }

// Store represents a content store.
type Store interface {
	Save(id, dir string) error
	Load(id, dir string) error

	SaveStream(id string, stream io.ReadCloser) error
	LoadStream(id string) (io.ReadCloser, error)

	PutKey(key, val string) error
	GetKey(key string) (string, error)
}

type local struct {
	dir string
}

func (s *local) Save(id, dir string) error {
	key := s.dir + "/store/content/" + id
	if err := os.MkdirAll(filepath.Dir(key), 0700); err != nil {
		return err
	}
	return fileutils.Tar(dir, key)
}

func (s *local) Load(id, dir string) error {
	os.MkdirAll(dir, 0700)
	return fileutils.Untar(s.dir+"/store/content/"+id, dir)
}

func (s *local) SaveStream(id string, stream io.ReadCloser) error {
	key := s.dir + "/store/content/" + id
	if err := os.MkdirAll(filepath.Dir(key), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(key, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return err
	}
	return fileutils.CopyStream(stream, f)
}

func (s *local) LoadStream(id string) (io.ReadCloser, error) {
	key := s.dir + "/store/content/" + id
	return os.Open(key)
}

func (s *local) PutKey(id, val string) error {
	key := s.dir + "/store/keys/" + id
	if err := os.MkdirAll(filepath.Dir(key), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(key, []byte(val), 0700)
}

func (s *local) GetKey(id string) (string, error) {
	key := s.dir + "/store/keys/" + id
	data, err := ioutil.ReadFile(key)
	return string(data), err
}
