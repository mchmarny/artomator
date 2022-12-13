package object

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
)

func Save(ctx context.Context, sha, bucket, dir string) error {
	if sha == "" || bucket == "" || dir == "" {
		return errors.New("sha, bucket, and dir are all required")
	}

	m, err := getFiles(sha, dir)
	if err != nil {
		return errors.Wrapf(err, "error getting files from: %s", dir)
	}

	for k, v := range m {
		d, err := os.ReadFile(k)
		if err != nil {
			return errors.Wrapf(err, "error reading content from: %s", k)
		}
		if err := Put(ctx, bucket, v, d); err != nil {
			return errors.Wrapf(err, "error writing content from: %s to:%s/%s",
				k, bucket, v)
		}
	}

	return nil
}

func getFiles(sha, dir string) (map[string]string, error) {
	if sha == "" || dir == "" {
		return nil, errors.New("sha, and dir are all required")
	}

	m := make(map[string]string)

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading files from dir: %s", dir)
	}

	for _, file := range files {
		m[path.Join(dir, file.Name())] = fmt.Sprintf("%s-%s", sha, file.Name())
	}

	return m, nil
}
