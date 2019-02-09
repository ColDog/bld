package store

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/coldog/bld/pkg/log"
	"github.com/moby/moby/client"
)

// ImageStore represents saving of a docker image.
type ImageStore interface {
	// Save will save a given image from the docker daemon.
	Save(ctx context.Context, name, id string) error
	// Restore will restore a given image from the docker repo.
	Restore(ctx context.Context, name, id string) error
}

// NewImageStore instantiates a new store given a store implementation. This
// will use docker load and docker save to store the images as tar archives in
// the provided store.
func NewImageStore(store Store) (ImageStore, error) {
	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &imageStore{
		store:  store,
		client: client,
	}, nil
}

type imageStore struct {
	store  Store
	client *client.Client
}

func (s *imageStore) Save(ctx context.Context, name, id string) error {
	log.ContextGetLogger(ctx).V(3).Printf("image store: saving image id=%s", id)

	imageID := name + ":" + id
	r, err := s.client.ImageSave(ctx, []string{imageID})
	if err != nil {
		return err
	}
	return s.store.SaveStream(id, r)
}

func (s *imageStore) Restore(ctx context.Context, name, id string) error {
	log.ContextGetLogger(ctx).V(3).Printf("image store: restoring image id=%s", id)

	r, err := s.store.LoadStream(id)
	if err != nil {
		return err
	}

	res, err := s.client.ImageLoad(ctx, r, true)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()
	return err
}
