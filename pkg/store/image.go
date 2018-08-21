package store

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
)

// ImageStore represents saving of a docker image.
type ImageStore interface {
	// Save will save a given image from the docker daemon.
	Save(ctx context.Context, name, id string) error
	// Restore will restore a given image from the docker repo..
	Restore(ctx context.Context, name, id string) error
}

// NewImageStore instantiates a new store.
func NewImageStore(repo, auth string) (ImageStore, error) {
	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &imageStore{
		repo:   repo,
		auth:   auth,
		client: client,
	}, nil
}

type imageStore struct {
	repo   string
	auth   string
	client *client.Client
}

func (s *imageStore) Save(ctx context.Context, name, id string) error {
	repoID := s.repo + "/" + name + ":" + id
	if err := s.client.ImageTag(ctx, name+":"+id, repoID); err != nil {
		return err
	}
	r, err := s.client.ImagePush(ctx, repoID, types.ImagePushOptions{
		RegistryAuth: s.auth,
	})
	if err != nil {
		return err
	}
	return read(ctx, r)
}

func (s *imageStore) Restore(ctx context.Context, name, id string) error {
	repoID := s.repo + "/" + name + ":" + id
	image := name + ":" + id
	r, perr := s.client.ImagePull(ctx, repoID, types.ImagePullOptions{
		RegistryAuth: s.auth,
	})
	if perr != nil {
		return perr
	}
	if err := read(ctx, r); err != nil {
		return err
	}
	return s.client.ImageTag(ctx, repoID, image)
}

func read(ctx context.Context, reader io.ReadCloser) error {
	defer reader.Close()
	scan := bufio.NewScanner(reader)
	for scan.Scan() {
		data := map[string]interface{}{}
		if err := json.Unmarshal(scan.Bytes(), &data); err != nil {
			return err
		}
		if err, ok := data["error"]; ok {
			return fmt.Errorf("%v", err)
		}
	}
	return nil
}
